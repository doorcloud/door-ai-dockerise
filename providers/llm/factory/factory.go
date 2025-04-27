package factory

import (
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/ollama"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

// New creates a new LLM provider based on the given name
func New(name string) (core.ChatCompletion, error) {
	switch name {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}
		return openai.New(apiKey), nil
	case "ollama":
		return ollama.New(), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
