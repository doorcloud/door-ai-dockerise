package factfix

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
)

// Fix attempts to fix invalid facts using the LLM
func Fix(ctx context.Context, client facts.LLMClient, facts *facts.Facts, buildDir string) error {
	// Set the build directory
	facts.BuildDir = buildDir

	// Convert facts to map for LLM
	factsMap := make(map[string]interface{})
	data, err := json.Marshal(facts)
	if err != nil {
		return fmt.Errorf("failed to marshal facts: %w", err)
	}
	if err := json.Unmarshal(data, &factsMap); err != nil {
		return fmt.Errorf("failed to unmarshal facts: %w", err)
	}

	// Generate new facts using LLM
	newFacts, err := client.GenerateFacts(ctx, []string{string(data)})
	if err != nil {
		return fmt.Errorf("failed to generate new facts: %w", err)
	}

	// Update facts with new values
	if lang, ok := newFacts["language"].(string); ok {
		facts.Language = lang
	}
	if framework, ok := newFacts["framework"].(string); ok {
		facts.Framework = framework
	}
	if version, ok := newFacts["version"].(string); ok {
		facts.Version = version
	}
	if buildTool, ok := newFacts["build_tool"].(string); ok {
		facts.BuildTool = buildTool
	}
	if buildCmd, ok := newFacts["build_cmd"].(string); ok {
		facts.BuildCmd = buildCmd
	}
	if artifact, ok := newFacts["artifact"].(string); ok {
		facts.Artifact = artifact
	}

	// Validate the updated facts
	if err := facts.ValidateBasic(); err != nil {
		return err
	}

	// Check if artifact exists
	if facts.BuildDir == "" {
		return fmt.Errorf("build directory not set")
	}

	artifactPath := filepath.Join(facts.BuildDir, facts.Artifact)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		return fmt.Errorf("artifact %s does not exist in build directory %s", facts.Artifact, facts.BuildDir)
	}

	return nil
}
