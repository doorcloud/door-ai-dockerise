package loop

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/aliou/dockerfile-gen/internal/detect"
	"github.com/aliou/dockerfile-gen/internal/dockerfile"
	"github.com/aliou/dockerfile-gen/internal/facts"
	"github.com/aliou/dockerfile-gen/internal/verify"
)

const maxRetries = 3

// Run executes the Dockerfile generation loop:
// 1. Infer facts about the project
// 2. Generate Dockerfile
// 3. Verify the Dockerfile
// 4. If verification fails, retry with error feedback
func Run(ctx context.Context, path fs.FS) (string, error) {
	// Get facts about the project
	rule := detect.Rule{
		Name: "spring-boot",
		Tool: "maven",
	}

	projectFacts, err := facts.Infer(ctx, path, rule)
	if err != nil {
		return "", fmt.Errorf("infer facts: %w", err)
	}

	var lastError string
	var lastDockerfile string
	for i := 0; i < maxRetries; i++ {
		// Generate Dockerfile with previous error feedback
		df, err := dockerfile.Generate(ctx, projectFacts, lastDockerfile, lastError)
		if err != nil {
			return "", fmt.Errorf("generate dockerfile: %w", err)
		}

		// Verify the Dockerfile
		err = verify.Verify(ctx, path, df)
		if err == nil {
			return df, nil // Success!
		}

		lastError = err.Error()
		lastDockerfile = df
	}

	return "", fmt.Errorf("failed after %d attempts: %s", maxRetries, lastError)
}
