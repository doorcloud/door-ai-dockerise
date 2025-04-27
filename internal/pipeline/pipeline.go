package pipeline

import (
	"context"
	"io/fs"
	"os"
	"sync"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Pipeline represents a Dockerfile generation pipeline
type Pipeline struct {
	detectors     []core.Detector
	factProviders []core.FactProvider
	generator     core.Generator
	verifier      core.Verifier
	maxAttempts   int
}

type Options struct {
	Detectors     []core.Detector
	FactProviders []core.FactProvider
	Generator     core.Generator
	Verifier      core.Verifier
	MaxAttempts   int
}

// New creates a new Pipeline instance
func New(opts Options) *Pipeline {
	return &Pipeline{
		detectors:     opts.Detectors,
		factProviders: opts.FactProviders,
		generator:     opts.Generator,
		verifier:      opts.Verifier,
		maxAttempts:   opts.MaxAttempts,
	}
}

// Process executes the pipeline
func (p *Pipeline) Process(ctx context.Context, dir string) (string, error) {
	fsys := os.DirFS(dir)

	// Detect stack
	stack, err := p.detectStack(ctx, fsys)
	if err != nil {
		return "", err
	}

	// Gather facts
	facts, err := p.gatherFacts(ctx, fsys, stack)
	if err != nil {
		return "", err
	}

	// Generate Dockerfile
	dockerfile, err := p.generateDockerfile(ctx, facts)
	if err != nil {
		return "", err
	}

	// Verify Dockerfile
	if p.verifier != nil {
		if err := p.verifyDockerfile(ctx, dir, dockerfile); err != nil {
			return "", err
		}
	}

	return dockerfile, nil
}

func (p *Pipeline) detectStack(ctx context.Context, fsys fs.FS) (core.StackInfo, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, len(p.detectors))
	stackCh := make(chan core.StackInfo, 1)

	for _, d := range p.detectors {
		wg.Add(1)
		go func(d core.Detector) {
			defer wg.Done()
			stack, err := d.Detect(ctx, fsys)
			if err != nil {
				errCh <- err
				return
			}
			if stack.Name != "" {
				select {
				case stackCh <- stack:
				default:
				}
			}
		}(d)
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(stackCh)
	}()

	select {
	case stack := <-stackCh:
		return stack, nil
	case err := <-errCh:
		return core.StackInfo{}, err
	case <-ctx.Done():
		return core.StackInfo{}, ctx.Err()
	}
}

func (p *Pipeline) gatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	facts := core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}

	for _, provider := range p.factProviders {
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
	return p.generator.Generate(ctx, facts)
}

func (p *Pipeline) verifyDockerfile(ctx context.Context, path string, dockerfile string) error {
	if p.verifier == nil {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- p.verifier.Verify(ctx, path, dockerfile)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
