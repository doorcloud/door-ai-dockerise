package generate

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/aliou/dockerfile-gen/internal/rules"
	"github.com/aliou/dockerfile-gen/pkg/dockerverify"
)

// Generate orchestrates the Dockerfile generation pipeline
func Generate(ctx context.Context, fsys fs.FS, cli llm.Client, maxRetries int, buildTimeout time.Duration) (string, error) {
	// Detect the stack
	rule, err := rules.Detect(ctx, fsys)
	if err != nil {
		return "", fmt.Errorf("failed to detect stack: %w", err)
	}
	if rule == nil {
		return "", fmt.Errorf("no matching stack rule found")
	}

	// Get relevant snippets
	snippets, err := rule.Snippets(ctx, fsys)
	if err != nil {
		return "", fmt.Errorf("failed to get snippets: %w", err)
	}

	// Convert snippets to facts
	facts := map[string]interface{}{
		"snippets": snippets,
	}

	// Analyze facts
	facts, err = cli.AnalyzeFacts(ctx, facts)
	if err != nil {
		return "", fmt.Errorf("failed to analyze facts: %w", err)
	}

	var dockerfile string
	for i := 0; i < maxRetries; i++ {
		// Generate Dockerfile
		dockerfile, err = cli.GenerateDockerfile(ctx, facts)
		if err != nil {
			return "", fmt.Errorf("failed to generate Dockerfile: %w", err)
		}

		// Verify the Dockerfile
		ok, errLog, err := dockerverify.Verify(ctx, fsys, dockerfile, buildTimeout)
		if err != nil {
			return "", fmt.Errorf("verification failed: %w", err)
		}
		if ok {
			return dockerfile, nil
		}

		// If verification failed, use the error log to generate a new Dockerfile
		facts["error_log"] = errLog
		dockerfile, err = cli.GenerateDockerfile(ctx, facts)
		if err != nil {
			return "", fmt.Errorf("failed to fix Dockerfile: %w", err)
		}
	}

	return "", fmt.Errorf("verification failed after %d attempts", maxRetries)
}
