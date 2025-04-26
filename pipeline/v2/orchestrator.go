package v2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Orchestrator coordinates the Dockerfile generation pipeline
type Orchestrator struct {
	detectors core.Detector
	llm       core.ChatCompletion
	verifier  core.Verifier
}

// New creates a new Orchestrator
func New(
	detectors core.Detector,
	llm core.ChatCompletion,
	verifier core.Verifier,
) *Orchestrator {
	return &Orchestrator{
		detectors: detectors,
		llm:       llm,
		verifier:  verifier,
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

	// Verify the Dockerfile
	if err := o.verifier.Verify(ctx, root, dockerfile.Content); err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(root, "Dockerfile"), []byte(dockerfile.Content), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %v", err)
	}

	return nil
}
