package v1

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers"
)

// Orchestrator implements the core.Orchestrator interface
type Orchestrator struct {
	detectors     []core.Detector
	factProviders []core.FactProvider
	generator     core.DockerfileGen
	verifier      core.Verifier
	log           core.Logger
	attempts      int
	buildTimeout  int
	builder       core.BuildDriver
}

// Opts contains options for creating a new Orchestrator
type Opts struct {
	Detectors    []core.Detector
	Facts        []core.FactProvider
	Generator    core.DockerfileGen
	Verifier     core.Verifier
	Log          core.Logger
	Attempts     int
	BuildTimeout int
	Builder      core.BuildDriver
}

// New creates a new Orchestrator
func New(opts Opts) *Orchestrator {
	// Set default timeout to 20 minutes if not specified
	if opts.BuildTimeout == 0 {
		opts.BuildTimeout = 20
	}

	// Set default builder to Docker engine if not specified
	if opts.Builder == nil {
		opts.Builder = drivers.Default()
	}

	return &Orchestrator{
		detectors:     opts.Detectors,
		factProviders: opts.Facts,
		generator:     opts.Generator,
		verifier:      opts.Verifier,
		log:           opts.Log,
		attempts:      opts.Attempts,
		buildTimeout:  opts.BuildTimeout,
		builder:       opts.Builder,
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
	o.logf("Starting Dockerfile generation for %s with build timeout of %d minutes", root, o.buildTimeout)

	// Create a context with the build timeout
	buildCtx, cancel := context.WithTimeout(ctx, time.Duration(o.buildTimeout)*time.Minute)
	defer cancel()

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

	// 3. Generate initial Dockerfile
	dockerfile, err := o.generator.Generate(ctx, facts)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}
	o.logf("Generated Dockerfile")

	// 4. Verify with retry
	var lastErr error
	for i := 0; i < o.attempts; i++ {
		// Check for context cancellation
		select {
		case <-buildCtx.Done():
			return "", buildCtx.Err()
		default:
		}

		// Build the image
		_, err := o.builder.Build(buildCtx, core.BuildInput{
			ContextTar: createContextTar(root), // TODO: Implement this function
			Dockerfile: dockerfile,
		}, logs)
		if err == nil {
			return dockerfile, nil
		}
		lastErr = err
		o.logf("Build failed (attempt %d/%d): %v", i+1, o.attempts, err)

		if i < o.attempts-1 {
			// Try to fix the Dockerfile
			newDockerfile, fixErr := o.generator.Fix(ctx, dockerfile, err.Error())
			if fixErr != nil {
				o.logf("Failed to fix Dockerfile: %v", fixErr)
				continue
			}
			dockerfile = newDockerfile
			o.logf("Fixed Dockerfile, retrying build")
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("build failed after %d attempts: %w", o.attempts, lastErr)
	}
	return "", fmt.Errorf("build failed after %d attempts", o.attempts)
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
			if logs != nil {
				fmt.Fprintf(logs, "Detected stack: %s\n", stack.Name)
			}
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
			if logs != nil {
				fmt.Fprintf(logs, "Warning: fact provider failed: %v\n", err)
			}
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
