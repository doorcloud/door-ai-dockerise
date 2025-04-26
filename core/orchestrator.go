package core

import (
	"context"
	"io"
	"os"
)

// Orchestrator coordinates the Dockerfile generation process
type Orchestrator interface {
	// Run executes the complete Dockerfile generation workflow:
	// 1. Detect stack type
	// 2. Gather facts
	// 3. Generate Dockerfile
	// 4. Verify result
	// Logs are streamed to the provided writer
	Run(
		ctx context.Context,
		root string,
		spec *Spec,
		logs io.Writer,
	) (string /*dockerfile*/, error)
}

// orchestrator implements the Orchestrator interface
type orchestrator struct {
	detector  Detector
	generator Generator
	verifier  Verifier
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(
	detector Detector,
	generator Generator,
	verifier Verifier,
) Orchestrator {
	return &orchestrator{
		detector:  detector,
		generator: generator,
		verifier:  verifier,
	}
}

// Run implements the Orchestrator interface
func (o *orchestrator) Run(
	ctx context.Context,
	root string,
	spec *Spec,
	logs io.Writer,
) (string, error) {
	var stack StackInfo
	var err error

	if spec != nil {
		// Use the provided spec
		stack = StackInfo{
			Name:      spec.Framework,
			BuildTool: spec.BuildTool,
			Version:   spec.Version,
		}
	} else {
		// Detect the stack
		fsys := os.DirFS(root)
		stack, err = o.detector.Detect(ctx, fsys)
		if err != nil {
			return "", err
		}
	}

	// Generate the Dockerfile
	dockerfile, err := o.generator.Generate(ctx, stack, nil)
	if err != nil {
		return "", err
	}

	// Verify the Dockerfile
	if err := o.verifier.Verify(ctx, root, dockerfile); err != nil {
		return "", err
	}

	return dockerfile, nil
}
