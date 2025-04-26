package v2

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

var (
	ErrNoStackDetected = errors.New("no stack detected")
)

// Orchestrator coordinates the Dockerfile generation pipeline
type Orchestrator struct {
	detector     core.Detector
	factProvider core.FactProvider
	generator    core.Generator
	dockerDriver docker.Driver
	maxRetries   int
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(
	detector core.Detector,
	factProvider core.FactProvider,
	generator core.Generator,
	dockerDriver docker.Driver,
	maxRetries int,
) *Orchestrator {
	return &Orchestrator{
		detector:     detector,
		factProvider: factProvider,
		generator:    generator,
		dockerDriver: dockerDriver,
		maxRetries:   maxRetries,
	}
}

// Run executes the Dockerfile generation pipeline
func (o *Orchestrator) Run(ctx context.Context, dir string) error {
	// Convert directory to filesystem
	fsys := os.DirFS(dir)

	// Detect stack type
	stack, err := o.detector.Detect(ctx, fsys)
	if err != nil {
		return err
	}

	// If no stack was detected, return an error
	if stack.Name == "" {
		return ErrNoStackDetected
	}

	// Gather facts about the stack
	facts, err := o.factProvider.Facts(ctx, stack)
	if err != nil {
		return err
	}

	// Generate Dockerfile
	dockerfile, err := o.generator.Generate(ctx, stack, facts)
	if err != nil {
		return err
	}

	// Write Dockerfile
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return err
	}

	// Build Docker image
	opts := docker.BuildOptions{
		Context:    dir,
		Tags:       []string{"myapp:latest"},
		Dockerfile: "Dockerfile",
	}
	if err := o.dockerDriver.Build(ctx, dockerfilePath, opts); err != nil {
		return err
	}

	return nil
}
