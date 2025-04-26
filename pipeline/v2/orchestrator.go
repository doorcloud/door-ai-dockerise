package v2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// Orchestrator coordinates the Dockerfile generation pipeline
type Orchestrator struct {
	detectors    core.Detector
	factProvider core.FactProvider
	llm          core.ChatCompletion
	builder      *docker.Driver
}

// New creates a new Orchestrator
func New(
	factProvider core.FactProvider,
	llm core.ChatCompletion,
) *Orchestrator {
	return &Orchestrator{
		detectors: core.DetectorChain{
			react.NewReactDetector(),
		},
		factProvider: factProvider,
		llm:          llm,
		builder:      docker.New(),
	}
}

// Run executes the Dockerfile generation pipeline
func (o *Orchestrator) Run(ctx context.Context, root string) error {
	// Detect the stack
	info, err := o.detectors.Detect(ctx, root)
	if err != nil {
		return fmt.Errorf("detection failed: %v", err)
	}
	if info.Name == "" {
		return fmt.Errorf("no stack detected")
	}

	// Gather facts about the stack
	facts, err := o.factProvider.Facts(ctx, info)
	if err != nil {
		return fmt.Errorf("failed to gather facts: %v", err)
	}

	// Generate Dockerfile using LLM
	dockerfile, err := o.llm.Chat(ctx, []core.Message{
		{
			Role:    "system",
			Content: "You are a Dockerfile expert. Generate a production-ready Dockerfile for the given stack.",
		},
		{
			Role:    "user",
			Content: info.Name,
		},
	})
	if err != nil {
		return fmt.Errorf("LLM failed: %v", err)
	}

	// Build and verify the Dockerfile
	imageID, err := o.builder.Build(ctx, root, dockerfile.Content)
	if err != nil {
		return fmt.Errorf("build failed: %v", err)
	}

	// Run the container to verify it works
	port := 3000
	if err := o.builder.Run(ctx, imageID, port); err != nil {
		return fmt.Errorf("run failed: %v", err)
	}

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(root, "Dockerfile"), []byte(dockerfile.Content), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %v", err)
	}

	return nil
}
