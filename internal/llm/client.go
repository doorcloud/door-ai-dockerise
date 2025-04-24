package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Interface defines the contract for LLM clients.
type Interface interface {
	GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error)
	FixDockerfile(ctx context.Context, facts map[string]interface{}, dockerfile string, buildDir string, errorLog string, errorType string, attempt int) (string, error)
	GenerateFacts(ctx context.Context, snippets []string) (map[string]interface{}, error)
}

// Client wraps the OpenAI client with our specific needs.
type Client struct {
	client *openai.Client
	logger *slog.Logger
}

// NewClient creates a new LLM client.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	client := openai.NewClient(apiKey)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// GenerateFacts prompts the LLM to extract facts from snippets.
func (c *Client) GenerateFacts(ctx context.Context, snippets []string) (map[string]interface{}, error) {
	prompt := buildFactsPrompt(snippets)

	if os.Getenv("DG_DEBUG") == "1" {
		c.logger.Debug("facts prompt", "content", prompt)
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	if os.Getenv("DG_DEBUG") == "1" {
		c.logger.Debug("facts response", "content", resp.Choices[0].Message.Content)
	}

	var facts map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &facts); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return facts, nil
}

// GenerateDockerfile creates a new Dockerfile based on the extracted facts.
func (c *Client) GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error) {
	prompt := buildDockerfilePrompt(facts)

	if os.Getenv("DG_DEBUG") == "1" {
		c.logger.Debug("dockerfile prompt", "content", prompt)
	}

	// Call OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	dockerfile := resp.Choices[0].Message.Content
	if os.Getenv("DG_DEBUG") == "1" {
		c.logger.Debug("generated dockerfile", "content", dockerfile)
	}

	return dockerfile, nil
}

// FixDockerfile attempts to fix a Dockerfile that failed to build.
func (c *Client) FixDockerfile(ctx context.Context, facts map[string]interface{}, dockerfile string, buildDir string, errorLog string, errorType string, attempt int) (string, error) {
	prompt := buildFixPrompt(facts, dockerfile, errorLog, errorType, attempt)

	if os.Getenv("DG_DEBUG") == "1" {
		c.logger.Debug("fix prompt", "content", prompt)
	}

	// Call OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	if os.Getenv("DG_DEBUG") == "1" {
		c.logger.Debug("fix response", "content", resp.Choices[0].Message.Content)
	}

	return resp.Choices[0].Message.Content, nil
}

// formatFacts formats the facts map for the prompt.
func formatFacts(facts map[string]interface{}) string {
	var sb strings.Builder
	for k, v := range facts {
		sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	return sb.String()
}

func buildFactsPrompt(snippets []string) string {
	return fmt.Sprintf(`SYSTEM
You are a code analysis expert. Given a set of code snippets, extract key facts about the project.
The output must be valid JSON with the following structure:
{
  "language": "primary language",
  "framework": "web framework if any",
  "version": "framework version if known",
  "build_tool": "build system (maven, gradle, etc)",
  "build_cmd": "command to build",
  "artifact": "path to built artifact",
  "ports": [list of ports],
  "env": ["list of env vars"],
  "health": "health check endpoint if any",
  "dependencies": ["list of key deps"]
}

SNIPPETS:
%s`, strings.Join(snippets, "\n\n"))
}

func buildDockerfilePrompt(facts map[string]interface{}) string {
	buildDir := facts["build_dir"].(string)

	// Get Maven version from env or use default
	mvnVersion := os.Getenv("DG_MVN_VERSION")
	if mvnVersion == "" {
		mvnVersion = "3.9.6"
	}

	var copyCmd string
	if buildDir != "." {
		copyCmd = fmt.Sprintf(`COPY %s/pom.xml .
COPY %s/.mvn .mvn
COPY %s/src src`, buildDir, buildDir, buildDir)
	} else {
		copyCmd = "COPY . ."
	}

	return fmt.Sprintf(`You are a Docker expert. Generate a Dockerfile for a Spring Boot application.
IMPORTANT: Reply ONLY with the raw Dockerfile content - no markdown, no explanations, no commentary.

Facts about the application:
%s

Use this template, replacing maven-%s with the correct version:
FROM eclipse-temurin:17-jdk AS build

WORKDIR /workspace

%s

# Use BuildKit cache for Maven dependencies
RUN --mount=type=cache,target=/root/.mvn \
    if [ -f "./mvnw" ]; then \
      chmod +x ./mvnw && \
      ./mvnw -q package -DskipTests; \
    else \
      curl -sL https://archive.apache.org/dist/maven/maven-3/%s/binaries/apache-maven-%s-bin.tar.gz | tar xz -C /tmp && \
      chmod +x /tmp/apache-maven-%s/bin/mvn && \
      ln -s /tmp/apache-maven-%s/bin/mvn /usr/bin/mvn && \
      mvn -q package -DskipTests; \
    fi

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=build /workspace/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, formatFacts(facts), mvnVersion, copyCmd, mvnVersion, mvnVersion, mvnVersion, mvnVersion)
}

func buildFixPrompt(facts map[string]interface{}, prevDockerfile string, errorLog string, errorType string, attempt int) string {
	// Get Maven version from env or use default
	mvnVersion := os.Getenv("DG_MVN_VERSION")
	if mvnVersion == "" {
		mvnVersion = "3.9.6"
	}

	return fmt.Sprintf(`You are a Docker expert. Fix the Dockerfile that failed to build.
IMPORTANT: Reply ONLY with the raw Dockerfile content - no markdown, no explanations, no commentary.

Error type: %s
Error log:
%s

Previous Dockerfile:
%s

Facts about the application:
%s

The build failed on attempt %d. Fix the Dockerfile while:
1. Using BuildKit cache for Maven dependencies
2. Skipping tests during the build
3. Using multi-stage build to minimize image size
4. Keeping the same base image unless absolutely necessary
5. Using Maven version %s in the fallback path`, errorType, errorLog, prevDockerfile, formatFacts(facts), attempt, mvnVersion)
}
