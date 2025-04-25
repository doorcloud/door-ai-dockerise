package dockerfile

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aliou/dockerfile-gen/internal/facts"
	"github.com/sashabaranov/go-openai"
)

// Generate creates a Dockerfile based on the provided facts
func Generate(ctx context.Context, facts facts.Facts, lastDockerfile string, lastError string) (string, error) {
	// Build the prompt for Dockerfile generation
	prompt := buildDockerfilePrompt(facts, lastDockerfile, lastError)

	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is required")
	}
	client := openai.NewClient(apiKey)

	// Call OpenAI to generate Dockerfile
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai call failed: %w", err)
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// buildDockerfilePrompt creates the prompt for Dockerfile generation
func buildDockerfilePrompt(facts facts.Facts, currentDF string, lastErr string) string {
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

	if currentDF != "" {
		prompt += fmt.Sprintf(`

Previous Dockerfile that failed:
%s

Error:
%s

Please fix the issues while maintaining the working parts.`, currentDF, lastErr)
	}

	return prompt
} 