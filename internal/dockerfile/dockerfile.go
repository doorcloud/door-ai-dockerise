package dockerfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// Generate creates a Dockerfile based on the provided facts
func Generate(ctx context.Context, facts types.Facts, client llm.Client) (string, error) {
	// Build the prompt for Dockerfile generation
	prompt := buildDockerfilePrompt(facts, "", "")

	// Call LLM to generate Dockerfile
	resp, err := client.Chat(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("llm call failed: %w", err)
	}

	return strings.TrimSpace(resp), nil
}

// buildDockerfilePrompt creates the prompt for Dockerfile generation
func buildDockerfilePrompt(facts types.Facts, currentDF string, lastErr string) string {
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