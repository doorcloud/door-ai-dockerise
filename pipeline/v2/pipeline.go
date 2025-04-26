package v2

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// Pipeline represents the v2 Dockerfile generation pipeline
type Pipeline struct {
	orchestrator *Orchestrator
}

// Option configures the pipeline
type Option func(*Pipeline)

// WithLLM sets the LLM provider for the pipeline
func WithLLM(llm core.ChatCompletion) Option {
	return func(p *Pipeline) {
		p.orchestrator.llm = llm
	}
}

// WithDockerDriver sets the Docker driver for the pipeline
func WithDockerDriver(driver docker.Driver) Option {
	return func(p *Pipeline) {
		p.orchestrator.builder = driver
	}
}

// NewPipeline creates a new v2 pipeline with the given options
func NewPipeline(opts ...Option) *Pipeline {
	p := &Pipeline{
		orchestrator: New(nil, nil), // We'll override these with options
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Run executes the pipeline on the given path
func (p *Pipeline) Run(ctx context.Context, path string) error {
	return p.orchestrator.Run(ctx, path)
}
