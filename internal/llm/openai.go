package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// openAIClient implements the Client interface using OpenAI's API
type openAIClient struct {
	client      *openai.Client
	model       string
	temperature float32
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &openAIClient{
		client:      openai.NewClient(apiKey),
		model:       "gpt-4",
		temperature: 0.7,
	}, nil
}

// AnalyzeFacts analyzes code snippets and returns enhanced facts
func (c *openAIClient) AnalyzeFacts(ctx context.Context, snippets []string) (Facts, error) {
	prompt := buildFactsPrompt(snippets)
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
		return Facts{}, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return Facts{}, fmt.Errorf("no completion choices returned")
	}

	var facts Facts
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &facts); err != nil {
		return Facts{}, fmt.Errorf("failed to parse facts: %w", err)
	}

	return facts, nil
}

// GenerateDockerfile generates a Dockerfile based on the facts
func (c *openAIClient) GenerateDockerfile(ctx context.Context, facts Facts, prevDockerfile string, prevError string) (string, error) {
	prompt := buildDockerfilePrompt(facts, prevDockerfile, prevError)
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
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// buildFactsPrompt builds the prompt for fact extraction
func buildFactsPrompt(snippets []string) string {
	return fmt.Sprintf(`You are a code analysis expert. Given a set of code snippets, extract key facts about the project.
The output must be valid JSON with the following structure:
{
  "language": "primary language",
  "framework": "web framework if any",
  "build_tool": "build system (maven, gradle, etc)",
  "build_cmd": "command to build",
  "build_dir": "directory containing build files (e.g., '.', 'backend/')",
  "start_cmd": "command to start the application",
  "artifact": "path to built artifact",
  "ports": [list of ports],
  "env": {"key": "value"},
  "health": "health check endpoint if any"
}

Code snippets:
%s`, snippets)
}

// buildDockerfilePrompt builds the prompt for Dockerfile generation
func buildDockerfilePrompt(facts Facts, prevDockerfile string, prevError string) string {
	prompt := fmt.Sprintf(`You are a Docker expert. Create a production-ready Dockerfile for a %s application using %s.
Facts about the application:
- Language: %s
- Framework: %s
- Build tool: %s
- Build command: %s
- Start command: %s
- Ports: %v
- Health check: %s
- Base image: %s

Requirements:
- Use multi-stage build
- Optimize for production
- Include health check
- Set appropriate labels
- Use non-root user
- Handle environment variables
- Include proper error handling

The Dockerfile should be valid and buildable.`, facts.Language, facts.Framework, facts.Language, facts.Framework,
		facts.BuildTool, facts.BuildCmd, facts.StartCmd, facts.Ports, facts.Health, facts.BaseImage)

	if prevDockerfile != "" {
		prompt += fmt.Sprintf(`

Previous Dockerfile that failed:
%s

Error:
%s

Please fix the issues while maintaining the working parts.`, prevDockerfile, prevError)
	}

	return prompt
}
