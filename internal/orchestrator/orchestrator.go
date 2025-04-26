package orchestrator

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

	// Cached results
	cachedStack core.StackInfo
	cachedFacts core.Facts
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
	} else if o.cachedStack.Name == "" {
		// Code-first mode: detect stack if not cached
		stack, err = o.detectStack(ctx, root, logs)
		if err != nil {
			return "", fmt.Errorf("detection failed: %w", err)
		}
		o.cachedStack = stack
	} else {
		// Use cached stack info
		stack = o.cachedStack
	}
	o.logf("Using stack: %s", stack.Name)

	// 2. Gather facts
	var facts core.Facts
	if o.cachedFacts.StackType == "" {
		facts, err = o.gatherFacts(ctx, root, stack, logs)
		if err != nil {
			return "", fmt.Errorf("fact gathering failed: %w", err)
		}
		o.cachedFacts = facts
	} else {
		facts = o.cachedFacts
	}
	o.logf("Using facts for stack: %s", stack.Name)

	// 3. Generate initial Dockerfile
	dockerfile, err := o.generator.Generate(ctx, facts)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}
	o.logf("Generated Dockerfile")

	// Create build input
	buildInput := core.BuildInput{
		ContextTar: createContextTar(root),
		Dockerfile: dockerfile,
	}

	// 4. Build with retry
	var lastErr error
	for i := 0; i < o.attempts; i++ {
		// Check for context cancellation
		select {
		case <-buildCtx.Done():
			return "", buildCtx.Err()
		default:
		}

		// Build the image
		_, err := o.builder.Build(buildCtx, buildInput, logs)
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
			buildInput.Dockerfile = newDockerfile
			o.logf("Fixed Dockerfile, retrying build")
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("build failed after %d attempts: %w", o.attempts, lastErr)
	}
	return "", fmt.Errorf("build failed after %d attempts", o.attempts)
}

// detectStack runs all detectors in parallel and returns the first successful result
func (o *Orchestrator) detectStack(ctx context.Context, root string, logs io.Writer) (core.StackInfo, error) {
	results := make(chan core.StackInfo, len(o.detectors))
	errors := make(chan error, len(o.detectors))

	// Convert root string to fs.FS
	fsys := os.DirFS(root)

	for _, detector := range o.detectors {
		go func(d core.Detector) {
			stack, err := d.Detect(ctx, fsys)
			if err != nil {
				errors <- err
				return
			}
			results <- stack
		}(detector)
	}

	// Wait for first successful result
	for i := 0; i < len(o.detectors); i++ {
		select {
		case stack := <-results:
			return stack, nil
		case <-errors:
			continue
		case <-ctx.Done():
			return core.StackInfo{}, ctx.Err()
		}
	}

	return core.StackInfo{}, fmt.Errorf("no detector succeeded")
}

// gatherFacts collects facts from all providers
func (o *Orchestrator) gatherFacts(ctx context.Context, root string, stack core.StackInfo, logs io.Writer) (core.Facts, error) {
	facts := core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}

	for _, provider := range o.factProviders {
		factSlice, err := provider.Facts(ctx, stack)
		if err != nil {
			o.logf("Failed to get facts from provider: %v", err)
			continue
		}
		// Convert []Fact to Facts struct
		for _, fact := range factSlice {
			if fact.Key == "StackType" {
				facts.StackType = fact.Value
			} else if fact.Key == "BuildTool" {
				facts.BuildTool = fact.Value
			}
		}
	}

	return facts, nil
}

// createContextTar creates a tar archive of the build context
func createContextTar(root string) io.Reader {
	// Create a pipe to stream the tar
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		// TODO: Implement tar creation
		// For now, just write an empty tar
		pw.Write([]byte{})
	}()
	return pr
}

// ... rest of the existing code ...
