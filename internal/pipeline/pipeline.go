package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/internal/dockerfile"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/verify"
)

// Run executes the full pipeline
func Run(repoPath string, client llm.Client) error {
	// Detect project type
	fsys := os.DirFS(repoPath)
	rule, err := rules.Detect(fsys)
	if err != nil {
		return fmt.Errorf("no rule matched: %v", err)
	}

	// Extract facts
	facts, err := rules.GetFacts(fsys, rule)
	if err != nil {
		return fmt.Errorf("failed to extract facts: %v", err)
	}

	// Generate Dockerfile
	dockerfilePath := filepath.Join(repoPath, "Dockerfile")
	df, err := dockerfile.Generate(context.Background(), facts, client)
	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %v", err)
	}

	// Write Dockerfile
	if err := os.WriteFile(dockerfilePath, []byte(df), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %v", err)
	}

	// Verify Dockerfile
	if err := verify.Verify(context.Background(), fsys, df); err != nil {
		return fmt.Errorf("failed to verify Dockerfile: %v", err)
	}

	return nil
}
