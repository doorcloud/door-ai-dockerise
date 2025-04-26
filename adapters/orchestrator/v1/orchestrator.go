package v1

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Orchestrator implements the core.Orchestrator interface
type Orchestrator struct {
	detectors     []core.Detector
	factProviders []core.FactProvider
	generator     core.ChatCompletion
	verifier      core.Verifier
	log           core.Logger
	attempts      int
}

// Opts contains options for creating a new Orchestrator
type Opts struct {
	Detectors []core.Detector
	Facts     []core.FactProvider
	Generator core.ChatCompletion
	Verifier  core.Verifier
	Log       core.Logger
	Attempts  int
}

// New creates a new Orchestrator
func New(opts Opts) *Orchestrator {
	return &Orchestrator{
		detectors:     opts.Detectors,
		factProviders: opts.Facts,
		generator:     opts.Generator,
		verifier:      opts.Verifier,
		log:           opts.Log,
		attempts:      opts.Attempts,
	}
}

func (o *Orchestrator) logf(format string, v ...any) {
	if o.log != nil {
		o.log.Printf(format, v...)
	}
}

// Run implements core.Orchestrator
func (o *Orchestrator) Run(
	ctx context.Context,
	root string,
	spec *core.Spec,
	logs io.Writer,
) (string, error) {
	o.logf("Starting Dockerfile generation for %s", root)

	// 1. Get stack info from spec or detect
	var stack core.StackInfo
	var err error

	if spec != nil {
		// Spec-first mode: trust user input
		stack = core.StackInfo{
			Name:      spec.Framework,
			BuildTool: spec.BuildTool,
			Version:   spec.Version,
		}
	} else {
		// Code-first mode: detect stack
		stack, err = o.detectStack(ctx, root, logs)
		if err != nil {
			return "", fmt.Errorf("detection failed: %w", err)
		}
	}
	o.logf("Detected stack: %s", stack.Name)

	// 2. Gather facts
	facts, err := o.gatherFacts(ctx, root, stack, logs)
	if err != nil {
		return "", fmt.Errorf("fact gathering failed: %w", err)
	}
	o.logf("Gathered facts for stack: %s", stack.Name)

	// 3. Generate Dockerfile
	dockerfile, err := o.generator.GenerateDockerfile(ctx, facts)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}
	o.logf("Generated Dockerfile")

	// 4. Verify with retry
	if err := o.verifyWithRetry(ctx, root, dockerfile, logs); err != nil {
		return "", fmt.Errorf("verification failed: %w", err)
	}
	o.logf("Verified Dockerfile")

	return dockerfile, nil
}

func (o *Orchestrator) detectStack(
	ctx context.Context,
	root string,
	logs io.Writer,
) (core.StackInfo, error) {
	o.logf("Detecting stack...")
	fsys := os.DirFS(root)
	for _, d := range o.detectors {
		stack, err := d.Detect(ctx, fsys)
		if err == nil && stack.Name != "" {
			fmt.Fprintf(logs, "Detected stack: %s\n", stack.Name)
			return stack, nil
		}
	}
	return core.StackInfo{}, fmt.Errorf("no stack detected")
}

func (o *Orchestrator) gatherFacts(
	ctx context.Context,
	root string,
	stack core.StackInfo,
	logs io.Writer,
) (core.Facts, error) {
	var allFacts core.Facts

	for _, p := range o.factProviders {
		facts, err := p.Facts(ctx, stack)
		if err != nil {
			fmt.Fprintf(logs, "Warning: fact provider failed: %v\n", err)
			continue
		}
		// Convert []Fact to Facts
		for _, fact := range facts {
			switch fact.Key {
			case "stack_type":
				allFacts.StackType = fact.Value
			case "build_tool":
				allFacts.BuildTool = fact.Value
			}
		}
	}

	return allFacts, nil
}

func (o *Orchestrator) verifyWithRetry(
	ctx context.Context,
	root string,
	dockerfile string,
	logs io.Writer,
) error {
	const maxRetries = 3
	const retryDelay = time.Second

	for i := 0; i < maxRetries; i++ {
		if err := o.verifier.Verify(ctx, root, dockerfile); err == nil {
			return nil
		}
		if i < maxRetries-1 {
			fmt.Fprintf(logs, "Verification failed, retrying...\n")
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("verification failed after %d retries", maxRetries)
}
