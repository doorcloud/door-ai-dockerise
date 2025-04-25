package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/internal/dockerfile"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/doorcloud/door-ai-dockerise/internal/verify"
)

// Run executes the full pipeline
func Run(repoPath string, client llm.Client) error {
	// Detect project type
	rule := rules.Detect(repoPath)
	if rule == nil {
		return fmt.Errorf("no rule matched")
	}

	// Extract facts
	factsMap := rule.Facts(repoPath)
	if factsMap == nil {
		return fmt.Errorf("failed to extract facts")
	}

	// Convert facts to types.Facts
	facts := types.Facts{
		Language:  factsMap["language"].(string),
		Framework: factsMap["framework"].(string),
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
	fsys := os.DirFS(repoPath)
	if err := verify.Verify(context.Background(), fsys, df); err != nil {
		return fmt.Errorf("failed to verify Dockerfile: %v", err)
	}

	return nil
}
