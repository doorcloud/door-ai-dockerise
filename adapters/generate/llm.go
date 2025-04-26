package generate

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm"
)

type LLM struct {
	client llm.Client
}

func NewLLM(client llm.Client) *LLM {
	return &LLM{
		client: client,
	}
}

func (l *LLM) Generate(ctx context.Context, stack core.StackInfo, facts []core.Fact) (string, error) {
	// Convert facts to a format the LLM can understand
	factStrings := make([]string, len(facts))
	for i, fact := range facts {
		factStrings[i] = fact.Key + ":" + fact.Value
	}

	// Add stack info as facts
	factStrings = append(factStrings, "stack:"+stack.Name)
	for k, v := range stack.Meta {
		factStrings = append(factStrings, k+":"+v)
	}

	// Generate Dockerfile using LLM
	return l.client.GenerateDockerfile(ctx, factStrings)
}
