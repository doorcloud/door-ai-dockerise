package facts

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

type Provider struct {
	llm *openai.Provider
}

func NewProvider(llm *openai.Provider) *Provider {
	return &Provider{llm: llm}
}

func (p *Provider) Facts(ctx context.Context, stack core.StackInfo) ([]core.Fact, error) {
	// TODO: Implement fact gathering using LLM
	return []core.Fact{
		{Key: "stack_type", Value: stack.Name},
		{Key: "build_tool", Value: stack.BuildTool},
	}, nil
}

func DefaultProviders(llm *openai.Provider) []core.FactProvider {
	return []core.FactProvider{NewProvider(llm)}
}
