package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

var ErrNoStackDetected = errors.New("no stack detected")

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
func (o *Orchestrator) Run(ctx context.Context, path string) error {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return err
	}

	fsys := os.DirFS(path)

	// Detect stack
	stack, found, err := o.detector.Detect(ctx, fsys, nil)
	if err != nil {
		return fmt.Errorf("failed to detect stack: %w", err)
	}
	if !found {
		return ErrNoStackDetected
	}

	// Check context after detection
	if err := ctx.Err(); err != nil {
		return err
	}

	// Gather facts about the stack
	factSlice, err := o.factProvider.Facts(ctx, stack)
	if err != nil {
		return err
	}

	// Check context after fact gathering
	if err := ctx.Err(); err != nil {
		return err
	}

	// Convert []Fact to Facts struct
	facts := core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}
	for _, fact := range factSlice {
		switch fact.Key {
		case "stack_type":
			facts.StackType = fact.Value
		case "build_tool":
			facts.BuildTool = fact.Value
		}
	}

	// Generate Dockerfile
	dockerfile, err := o.generator.Generate(ctx, facts)
	if err != nil {
		return err
	}

	// Check context after generation
	if err := ctx.Err(); err != nil {
		return err
	}

	// Write Dockerfile
	dockerfilePath := filepath.Join(path, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0o644); err != nil {
		return err
	}

	// Build Docker image
	opts := docker.BuildOptions{
		Context:    path,
		Tags:       []string{"myapp:latest"},
		Dockerfile: "Dockerfile",
	}
	if err := o.dockerDriver.Build(ctx, dockerfilePath, opts); err != nil {
		return err
	}

	return nil
}
