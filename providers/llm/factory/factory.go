package factory

import (
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

// New creates a new LLM provider based on the given name
func New(name string) (*openai.Provider, error) {
	switch name {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}
		return openai.NewProvider(apiKey), nil
	case "ollama":
		return nil, fmt.Errorf("ollama provider not supported yet")
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
