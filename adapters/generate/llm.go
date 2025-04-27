package generate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// LLMGenerator generates Dockerfiles using LLM
type LLMGenerator struct {
	llm core.ChatCompletion
}

// NewLLM creates a new LLM generator
func NewLLM(llm core.ChatCompletion) *LLMGenerator {
	return &LLMGenerator{
		llm: llm,
	}
}

// GatherFacts implements the ChatCompletion interface
func (g *LLMGenerator) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return g.llm.GatherFacts(ctx, fsys, stack)
}

// Generate creates a Dockerfile using the provided facts
func (g *LLMGenerator) Generate(ctx context.Context, facts core.Facts) (string, error) {
	return g.llm.GenerateDockerfile(ctx, facts)
}

// GenerateDockerfile implements the ChatCompletion interface
func (g *LLMGenerator) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	return g.llm.GenerateDockerfile(ctx, facts)
}

func (g *LLMGenerator) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	// Convert error and previous Dockerfile to messages
	messages := []core.Message{
		{
			Role:    "system",
			Content: "Fix the Dockerfile based on the build error.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Previous Dockerfile:\n%s\n\nBuild error:\n%s", prevDockerfile, buildErr),
		},
	}

	// Get completion from LLM
	response, err := g.llm.Complete(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to fix Dockerfile: %w", err)
	}

	return response, nil
}
