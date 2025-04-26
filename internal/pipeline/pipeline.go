package pipeline

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Pipeline represents a Dockerfile generation pipeline
type Pipeline struct {
	detectors  []core.Detector
	generators []core.Generator
	verifiers  []core.Verifier
	providers  []core.FactProvider
}

// New creates a new Pipeline instance
func New(detectors []core.Detector, generators []core.Generator, verifiers []core.Verifier, providers []core.FactProvider) *Pipeline {
	return &Pipeline{
		detectors:  detectors,
		generators: generators,
		verifiers:  verifiers,
		providers:  providers,
	}
}

// Run executes the pipeline
func (p *Pipeline) Run(ctx context.Context, dir string, opts map[string]interface{}, streamer io.Writer) (string, error) {
	// Detect stack
	stack, err := p.detectStack(ctx, dir)
	if err != nil {
		return "", fmt.Errorf("failed to detect stack: %w", err)
	}

	// Gather facts
	facts, err := p.gatherFacts(ctx, stack)
	if err != nil {
		return "", fmt.Errorf("failed to gather facts: %w", err)
	}

	// Generate Dockerfile
	dockerfile, err := p.generateDockerfile(ctx, facts)
	if err != nil {
		return "", fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Verify Dockerfile
	if err := p.verifyDockerfile(ctx, dir, dockerfile); err != nil {
		return "", fmt.Errorf("failed to verify Dockerfile: %w", err)
	}

	return dockerfile, nil
}

func (p *Pipeline) detectStack(ctx context.Context, dir string) (core.StackInfo, error) {
	fsys := os.DirFS(dir)
	results := make(chan core.StackInfo, len(p.detectors))
	errs := make(chan error, len(p.detectors))

	var wg sync.WaitGroup
	for _, d := range p.detectors {
		wg.Add(1)
		go func(d core.Detector) {
			defer wg.Done()
			stack, err := d.Detect(ctx, fsys)
			if err != nil {
				errs <- err
				return
			}
			results <- stack
		}(d)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// Return first successful result
	for stack := range results {
		if stack.Name != "" {
			return stack, nil
		}
	}

	// If no successful results, return first error
	if err := <-errs; err != nil {
		return core.StackInfo{}, err
	}

	return core.StackInfo{}, fmt.Errorf("no stack detected")
}

func (p *Pipeline) gatherFacts(ctx context.Context, stack core.StackInfo) (core.Facts, error) {
	facts := core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}

	for _, provider := range p.providers {
		factSlice, err := provider.Facts(ctx, stack)
		if err != nil {
			return core.Facts{}, err
		}

		for _, fact := range factSlice {
			switch fact.Key {
			case "stack_type":
				facts.StackType = fact.Value
			case "build_tool":
				facts.BuildTool = fact.Value
			}
		}
	}

	return facts, nil
}

func (p *Pipeline) generateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	results := make(chan string, len(p.generators))
	errs := make(chan error, len(p.generators))

	var wg sync.WaitGroup
	for _, g := range p.generators {
		wg.Add(1)
		go func(g core.Generator) {
			defer wg.Done()
			dockerfile, err := g.Generate(ctx, facts)
			if err != nil {
				errs <- err
				return
			}
			results <- dockerfile
		}(g)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// Return first successful result
	for dockerfile := range results {
		if dockerfile != "" {
			return dockerfile, nil
		}
	}

	// If no successful results, return first error
	if err := <-errs; err != nil {
		return "", err
	}

	return "", fmt.Errorf("no Dockerfile generated")
}

func (p *Pipeline) verifyDockerfile(ctx context.Context, dir string, dockerfile string) error {
	errs := make(chan error, len(p.verifiers))

	var wg sync.WaitGroup
	for _, v := range p.verifiers {
		wg.Add(1)
		go func(v core.Verifier) {
			defer wg.Done()
			if err := v.Verify(ctx, dir, dockerfile); err != nil {
				errs <- err
			}
		}(v)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	// Return first error if any
	if err := <-errs; err != nil {
		return err
	}

	return nil
}
