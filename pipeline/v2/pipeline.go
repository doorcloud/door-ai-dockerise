package v2

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// Options configures the pipeline
type Options struct {
	Detectors     []core.Detector
	FactProviders []core.FactProvider
	Generator     core.Generator
	Verifier      docker.Driver
	MaxRetries    int
}

// Pipeline represents the v2 Dockerfile generation pipeline
type Pipeline struct {
	orchestrator *Orchestrator
}

// New creates a new v2 pipeline with the given options
func New(opts Options) *Pipeline {
	// Create detector chain
	detector := core.DetectorChain(opts.Detectors)

	// Create fact provider chain
	factProvider := FactProviderChain(opts.FactProviders)

	// Create orchestrator
	orchestrator := NewOrchestrator(detector, factProvider, opts.Generator, opts.Verifier, opts.MaxRetries)

	return &Pipeline{
		orchestrator: orchestrator,
	}
}

// Run executes the pipeline on the given path
func (p *Pipeline) Run(ctx context.Context, path string) error {
	return p.orchestrator.Run(ctx, path)
}

// FactProviderChain implements FactProvider by trying each provider in sequence
type FactProviderChain []core.FactProvider

// Facts implements the FactProvider interface for FactProviderChain
func (c FactProviderChain) Facts(ctx context.Context, stack core.StackInfo) ([]core.Fact, error) {
	var facts []core.Fact
	for _, p := range c {
		providerFacts, err := p.Facts(ctx, stack)
		if err != nil {
			return nil, err
		}
		facts = append(facts, providerFacts...)
	}
	return facts, nil
}

// NewPipeline creates a new pipeline with the given options
func NewPipeline(opts ...func(*Options)) *Pipeline {
	options := &Options{
		MaxRetries: 3,
	}

	for _, opt := range opts {
		opt(options)
	}

	return New(*options)
}

// WithDetectors sets the detectors for the pipeline
func WithDetectors(detectors ...core.Detector) func(*Options) {
	return func(opts *Options) {
		opts.Detectors = detectors
	}
}

// WithFactProviders sets the fact providers for the pipeline
func WithFactProviders(providers ...core.FactProvider) func(*Options) {
	return func(opts *Options) {
		opts.FactProviders = providers
	}
}

// WithGenerator sets the generator for the pipeline
func WithGenerator(generator core.Generator) func(*Options) {
	return func(opts *Options) {
		opts.Generator = generator
	}
}

// WithDockerDriver sets the Docker driver for the pipeline
func WithDockerDriver(driver docker.Driver) func(*Options) {
	return func(opts *Options) {
		opts.Verifier = driver
	}
}

// WithMaxRetries sets the maximum number of retries for the pipeline
func WithMaxRetries(maxRetries int) func(*Options) {
	return func(opts *Options) {
		opts.MaxRetries = maxRetries
	}
}
