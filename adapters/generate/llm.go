package generate

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// LLMGenerator implements the Generator interface using an LLM
type LLMGenerator struct {
	llm core.ChatCompletion
}

// NewLLM creates a new LLM-based generator
func NewLLM(llm core.ChatCompletion) *LLMGenerator {
	return &LLMGenerator{llm: llm}
}

// GatherFacts implements the ChatCompletion interface
func (g *LLMGenerator) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return g.llm.GatherFacts(ctx, fsys, stack)
}

// Generate implements the Generator interface
func (g *LLMGenerator) Generate(ctx context.Context, stack core.StackInfo, facts []core.Fact) (string, error) {
	// Convert facts to a format suitable for the LLM
	llmFacts := core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}

	// Generate Dockerfile using LLM
	return g.llm.GenerateDockerfile(ctx, llmFacts)
}

// GenerateDockerfile implements the ChatCompletion interface
func (g *LLMGenerator) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	return g.llm.GenerateDockerfile(ctx, facts)
}
