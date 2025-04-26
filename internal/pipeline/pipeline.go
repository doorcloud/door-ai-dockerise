package pipeline

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Pipeline represents the main processing pipeline
type Pipeline struct {
	detectors    []core.Detector
	generators   []core.Generator
	verifiers    []core.Verifier
	factProvider core.FactProvider
}

// NewPipeline creates a new pipeline instance
func NewPipeline(
	detectors []core.Detector,
	generators []core.Generator,
	verifiers []core.Verifier,
	factProvider core.FactProvider,
) *Pipeline {
	return &Pipeline{
		detectors:    detectors,
		generators:   generators,
		verifiers:    verifiers,
		factProvider: factProvider,
	}
}

// Process runs the complete pipeline
func (p *Pipeline) Process(ctx context.Context, input string) (*core.StackInfo, error) {
	// Step 1: Detect stack information
	stackInfo, err := p.detectStack(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("detection failed: %w", err)
	}

	// Step 2: Generate Dockerfile
	dockerfile, err := p.generateDockerfile(ctx, stackInfo)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// Step 3: Verify the generated Dockerfile
	if err := p.verifyDockerfile(ctx, dockerfile); err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	return stackInfo, nil
}

// detectStack runs all detectors in parallel and returns the first successful detection
func (p *Pipeline) detectStack(ctx context.Context, input string) (*core.StackInfo, error) {
	var wg sync.WaitGroup
	results := make(chan *core.StackInfo, len(p.detectors))
	errs := make(chan error, len(p.detectors))

	fsys := os.DirFS(input)
	for _, detector := range p.detectors {
		wg.Add(1)
		go func(d core.Detector) {
			defer wg.Done()
			if stack, err := d.Detect(ctx, fsys); err == nil {
				results <- &stack
			} else {
				errs <- err
			}
		}(detector)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// Return the first successful detection
	for stack := range results {
		return stack, nil
	}

	// If we get here, no detection was successful
	var allErrs []error
	for err := range errs {
		allErrs = append(allErrs, err)
	}
	return nil, fmt.Errorf("all detectors failed: %v", allErrs)
}

// generateDockerfile runs all generators in parallel and returns the first successful generation
func (p *Pipeline) generateDockerfile(ctx context.Context, stack *core.StackInfo) (string, error) {
	var wg sync.WaitGroup
	results := make(chan string, len(p.generators))
	errs := make(chan error, len(p.generators))

	facts, err := p.factProvider.Facts(ctx, *stack)
	if err != nil {
		return "", err
	}

	for _, generator := range p.generators {
		wg.Add(1)
		go func(g core.Generator) {
			defer wg.Done()
			if dockerfile, err := g.Generate(ctx, *stack, facts); err == nil {
				results <- dockerfile
			} else {
				errs <- err
			}
		}(generator)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// Return the first successful generation
	for dockerfile := range results {
		return dockerfile, nil
	}

	// If we get here, no generation was successful
	var allErrs []error
	for err := range errs {
		allErrs = append(allErrs, err)
	}
	return "", fmt.Errorf("all generators failed: %v", allErrs)
}

// verifyDockerfile runs all verifiers in parallel and returns the first error if any
func (p *Pipeline) verifyDockerfile(ctx context.Context, dockerfile string) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(p.verifiers))

	for _, verifier := range p.verifiers {
		wg.Add(1)
		go func(v core.Verifier) {
			defer wg.Done()
			if err := v.Verify(ctx, ".", dockerfile); err != nil {
				errs <- err
			}
		}(verifier)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	// Return the first error if any
	for err := range errs {
		return err
	}

	return nil
}
