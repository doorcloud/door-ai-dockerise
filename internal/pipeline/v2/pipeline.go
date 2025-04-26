package v2

import (
	"context"
	"fmt"
	"time"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type Pipeline struct {
	Detectors     []core.Detector
	FactProviders []core.FactProvider
	Generator     core.DockerfileGen
	Verifier      core.Verifier
	MaxAttempts   int
}

func NewPipeline(
	detectors []core.Detector,
	factProviders []core.FactProvider,
	generator core.DockerfileGen,
	verifier core.Verifier,
) *Pipeline {
	return &Pipeline{
		Detectors:     detectors,
		FactProviders: factProviders,
		Generator:     generator,
		Verifier:      verifier,
		MaxAttempts:   3,
	}
}

func (p *Pipeline) Run(ctx context.Context, repoPath string) (string, error) {
	// Step 1: Detect the stack
	var stack core.StackInfo
	var err error
	for _, detector := range p.Detectors {
		stack, err = detector.Detect(ctx, repoPath)
		if err == nil && stack.Name != "" {
			break
		}
	}
	if err != nil {
		return "", fmt.Errorf("detection failed: %w", err)
	}
	if stack.Name == "" {
		return "", fmt.Errorf("no stack detected")
	}

	// Step 2: Collect facts
	var allFacts []string
	for _, provider := range p.FactProviders {
		facts, err := provider.Facts(ctx, stack)
		if err != nil {
			return "", fmt.Errorf("facts collection failed: %w", err)
		}
		allFacts = append(allFacts, facts...)
	}

	// Step 3: Generate Dockerfile
	dockerfile, err := p.Generator.Generate(ctx, allFacts)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	// Step 4: Verify with retries
	var lastErr error
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		err = p.Verifier.Verify(ctx, repoPath, dockerfile)
		if err == nil {
			return dockerfile, nil
		}
		lastErr = err

		// Wait before retrying (exponential backoff)
		if attempt < p.MaxAttempts {
			waitTime := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(waitTime):
				continue
			}
		}
	}

	return "", fmt.Errorf("verification failed after %d attempts: %w", p.MaxAttempts, lastErr)
}
