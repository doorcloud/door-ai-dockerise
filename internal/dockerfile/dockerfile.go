package dockerfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliou/dockerfile-gen/internal/facts"
)

// Generate creates a Dockerfile based on the provided facts
func Generate(ctx context.Context, facts facts.Facts) (string, error) {
	// Build the prompt for Dockerfile generation
	prompt := buildDockerfilePrompt(facts, "", "")

	// For now, return a basic Dockerfile template
	// This will be replaced with actual LLM call later
	dockerfile := fmt.Sprintf(`FROM %s AS builder
WORKDIR /app
COPY . .
RUN %s

FROM %s
WORKDIR /app
COPY --from=builder /app/%s .
EXPOSE %d
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:%d%s || exit 1
CMD %s`, 
		facts.BaseImage, 
		facts.BuildCmd,
		facts.BaseImage,
		facts.Artifact,
		facts.Ports[0],
		facts.Ports[0],
		facts.Health,
		facts.StartCmd)

	return strings.TrimSpace(dockerfile), nil
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