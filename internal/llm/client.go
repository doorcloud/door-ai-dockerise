package llm

import (
	"context"
	"fmt"
	"os"
)

// Facts represents the analyzed facts about a project
type Facts struct {
	Language  string
	Framework string
	BuildTool string
	BuildCmd  string
	StartCmd  string
	Ports     []int
	Health    string
	BaseImage string
	Env       map[string]string
	Artifact  string
	BuildDir  string
}

// Client defines the interface for LLM clients
type Client interface {
	// AnalyzeFacts analyzes facts about a technology stack
	AnalyzeFacts(ctx context.Context, facts map[string]interface{}) (map[string]interface{}, error)

	// GenerateDockerfile generates a Dockerfile based on analyzed facts
	GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error)
}

// NewClient creates a new LLM client
func NewClient() (Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	return NewOpenAIClient(apiKey), nil
}
