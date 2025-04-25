package loop

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/dockerfile"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/doorcloud/door-ai-dockerise/internal/verify"
)

// Run executes the Dockerfile generation loop
func Run(ctx context.Context, fsys fs.FS, client llm.Client) (string, error) {
	// Detect project type
	rule, err := detect.Detect(fsys)
	if err != nil {
		return "", fmt.Errorf("detect project: %w", err)
	}

	// Infer facts about the project
	projectFacts, err := facts.InferWithClient(ctx, fsys, rule, client)
	if err != nil {
		return "", fmt.Errorf("infer facts: %w", err)
	}

	// Convert facts to types.Facts
	typedFacts := types.Facts{
		Language:  projectFacts.Language,
		Framework: projectFacts.Framework,
		BuildTool: projectFacts.BuildTool,
		BuildCmd:  projectFacts.BuildCmd,
		BuildDir:  projectFacts.BuildDir,
		StartCmd:  projectFacts.StartCmd,
		Artifact:  projectFacts.Artifact,
		Ports:     projectFacts.Ports,
		Health:    projectFacts.Health,
		Env:       projectFacts.Env,
		BaseImage: projectFacts.BaseImage,
	}

	// Try generating and verifying up to 3 times
	var lastError error
	for i := 0; i < 3; i++ {
		// Generate Dockerfile
		df, err := dockerfile.Generate(ctx, typedFacts, client)
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
