package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/config"
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
	// client is the OpenAI client instance
	client *openai.Client
	// logger is used for application logging
	logger *slog.Logger
	// model is the OpenAI model to use (e.g., "gpt-4")
	model string
	// temperature controls the randomness of the model's output (0.0 to 2.0)
	temperature float32
	// config holds the application configuration
	config *config.Config
}

// NewClient creates a new LLM client with configuration from environment variables.
// It validates and sets up the client with appropriate defaults.
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg.OpenAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	// Validate model name
	if !isValidModel(cfg.OpenAIModel) {
		return nil, fmt.Errorf("invalid model: %s", cfg.OpenAIModel)
	}

	// Validate temperature
	if cfg.OpenAITemp < 0.0 || cfg.OpenAITemp > 2.0 {
		return nil, fmt.Errorf("temperature must be between 0.0 and 2.0, got: %f", cfg.OpenAITemp)
	}

	// Get log level from config
	logLevel := slog.LevelDebug
	switch strings.ToLower(cfg.OpenAILogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	client := openai.NewClient(cfg.OpenAIKey)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return &Client{
		client:      client,
		logger:      logger,
		model:       cfg.OpenAIModel,
		temperature: float32(cfg.OpenAITemp),
		config:      cfg,
	}, nil
}

// isValidModel checks if the model name is valid
func isValidModel(model string) bool {
	validModels := map[string]bool{
		"gpt-4":         true,
		"gpt-4-turbo":   true,
		"gpt-3.5-turbo": true,
		"gpt-4.1-mini":  true,
	}
	return validModels[model]
}

// GenerateFacts prompts the LLM to extract facts from snippets.
func (c *Client) GenerateFacts(ctx context.Context, snippets []string) (map[string]interface{}, error) {
	prompt := buildFactsPrompt(snippets)

	if c.config.Debug {
		c.logger.Debug("facts prompt", "content", prompt)
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
			},
			Temperature: c.temperature,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	if c.config.Debug {
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
	prompt := buildDockerfilePrompt(facts, c.config)

	if c.config.Debug {
		c.logger.Debug("dockerfile prompt", "content", prompt)
	}

	// Call OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	dockerfile := resp.Choices[0].Message.Content
	if c.config.Debug {
		c.logger.Debug("generated dockerfile", "content", dockerfile)
	}

	return dockerfile, nil
}

// FixDockerfile attempts to fix a Dockerfile that failed to build.
func (c *Client) FixDockerfile(ctx context.Context, facts map[string]interface{}, dockerfile string, buildDir string, errorLog string, errorType string, attempt int) (string, error) {
	prompt := buildFixPrompt(facts, dockerfile, errorLog, errorType, attempt)

	if c.config.Debug {
		c.logger.Debug("fix prompt", "content", prompt)
	}

	// Call OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	if c.config.Debug {
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
  "build_dir": "directory containing build files (e.g., '.', 'backend/')",
  "start_cmd": "command to start the application",
  "artifact": "path to built artifact",
  "ports": [list of ports],
  "env": {"key": "value"},
  "health": "health check endpoint if any",
  "dependencies": ["list of key deps"],
  "base_hint": "base image hint"
}

SNIPPETS:
%s`, strings.Join(snippets, "\n\n"))
}

func buildDockerfilePrompt(facts map[string]interface{}, cfg *config.Config) string {
	// Use Maven version from config
	mvnVersion := cfg.MvnVersion

	return fmt.Sprintf(`SYSTEM
You are a Docker expert. Create a production-ready Dockerfile for a Spring Boot application.
Use the following facts:
%s

Requirements:
- Use multi-stage build
- Use Maven %s
- Optimize for production
- Include health check
- Set appropriate labels
- Use non-root user
- Handle environment variables
- Include proper error handling

The output should be a valid Dockerfile.`, formatFacts(facts), mvnVersion)
}

func buildFixPrompt(facts map[string]interface{}, prevDockerfile string, errorLog string, errorType string, attempt int) string {
	return fmt.Sprintf(`SYSTEM
You are a Docker expert. Fix the following Dockerfile that failed to build.

Error type: %s
Attempt: %d
Error log:
%s

Previous Dockerfile:
%s

Project facts:
%s

Requirements:
- Fix the specific error
- Maintain all working parts
- Explain the fix in a comment
- Keep the multi-stage build
- Preserve security best practices

The output should be a valid Dockerfile.`, errorType, attempt, errorLog, prevDockerfile, formatFacts(facts))
}
