package ollama

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Client implements the ChatCompletion interface for Ollama
type Client struct{}

// New creates a new Ollama client
func New() *Client {
	return &Client{}
}

// Complete implements ChatCompletion
func (c *Client) Complete(ctx context.Context, messages []core.Message) (string, error) {
	// For now, just return a basic Dockerfile
	return "FROM ubuntu:latest\n", nil
}

// GatherFacts implements ChatCompletion
func (c *Client) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return core.Facts{}, nil
}

// GenerateDockerfile implements ChatCompletion
func (c *Client) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	return "FROM ubuntu:latest\n", nil
}
