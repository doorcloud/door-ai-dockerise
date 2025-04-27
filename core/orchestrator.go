package core

import (
	"context"
	"fmt"
	"os"
)

// Orchestrator coordinates the detection, generation, and verification of Dockerfiles
type Orchestrator struct {
	detector  Detector
	generator Generator
	verifier  Verifier
}

// NewOrchestrator creates a new Orchestrator
func NewOrchestrator(detector Detector, generator Generator, verifier Verifier) *Orchestrator {
	return &Orchestrator{
		detector:  detector,
		generator: generator,
		verifier:  verifier,
	}
}

// Run orchestrates the Dockerfile generation process
func (o *Orchestrator) Run(ctx context.Context, root string, spec *Spec) (string, error) {
	// 1. Detect stack
	var stack StackInfo
	var err error
	if spec != nil {
		stack = StackInfo{
			Name:      spec.Framework,
			BuildTool: spec.BuildTool,
			Version:   spec.Version,
		}
	} else {
		fsys := os.DirFS(root)
		stack, _, err = o.detector.Detect(ctx, fsys, nil)
		if err != nil {
			return "", fmt.Errorf("failed to detect stack: %w", err)
		}
	}

	// 2. Generate Dockerfile
	facts := Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}
	dockerfile, err := o.generator.Generate(ctx, facts)
	if err != nil {
		return "", fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// 3. Verify Dockerfile
	if err := o.verifier.Verify(ctx, root, dockerfile); err != nil {
		// Try to fix the Dockerfile
		newDockerfile, fixErr := o.generator.Fix(ctx, dockerfile, err.Error())
		if fixErr != nil {
			return "", fmt.Errorf("failed to fix Dockerfile: %w", fixErr)
		}
		dockerfile = newDockerfile

		// Verify the fixed Dockerfile
		if err := o.verifier.Verify(ctx, root, dockerfile); err != nil {
			return "", fmt.Errorf("failed to verify fixed Dockerfile: %w", err)
		}
	}

	return dockerfile, nil
}
