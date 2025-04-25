package loop

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/aliou/dockerfile-gen/internal/detect"
	"github.com/aliou/dockerfile-gen/internal/dockerfile"
	"github.com/aliou/dockerfile-gen/internal/facts"
	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/aliou/dockerfile-gen/internal/verify"
)

// Run executes the Dockerfile generation loop
func Run(ctx context.Context, fsys fs.FS, client llm.Client) (string, error) {
	// Detect project type
	rule, err := detect.Detect(fsys)
	if err != nil {
		return "", fmt.Errorf("detect project: %w", err)
	}

	// Infer facts about the project
	facts, err := facts.InferWithClient(ctx, fsys, rule, client)
	if err != nil {
		return "", fmt.Errorf("infer facts: %w", err)
	}

	// Try generating and verifying up to 3 times
	var lastError error
	for i := 0; i < 3; i++ {
		// Generate Dockerfile
		df, err := dockerfile.Generate(ctx, facts, client)
		if err != nil {
			return "", fmt.Errorf("generate dockerfile: %w", err)
		}

		// Verify the Dockerfile
		if err := verify.Verify(ctx, fsys, df); err == nil {
			return df, nil // Success!
		} else {
			lastError = err
		}
	}

	return "", fmt.Errorf("failed after 3 attempts: %v", lastError)
}
