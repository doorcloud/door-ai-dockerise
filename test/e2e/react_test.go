package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/node"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers/docker"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

func TestReactE2E(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test. Set DG_E2E=1 to run.")
	}

	// Get the absolute path to the fixtures directory
	fixturesDir := filepath.Join("test", "e2e", "fixtures", "react-min")
	absPath, err := filepath.Abs(fixturesDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create the pipeline with mock LLM
	p := pipeline.New(
		core.DetectorChain{
			react.New(),
			node.New(),
		},
		mock.NewMockLLM(),
		docker.New(),
	)

	// Run the pipeline
	if err := p.Run(ctx, absPath); err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Verify Dockerfile exists
	dockerfilePath := filepath.Join(absPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Fatalf("Dockerfile was not generated")
	}

	// Clean up
	if err := os.Remove(dockerfilePath); err != nil {
		t.Logf("Warning: failed to clean up Dockerfile: %v", err)
	}
}
