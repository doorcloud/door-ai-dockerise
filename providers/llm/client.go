package llm

import "context"

// Client defines the interface for LLM providers
type Client interface {
	GenerateDockerfile(ctx context.Context, facts []string) (string, error)
}
