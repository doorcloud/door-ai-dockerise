package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

// RunTestWithProject runs the pipeline on a test project
func RunTestWithProject(t *testing.T, projectPath string) {
	ctx := context.Background()

	// Create mock implementations
	mockLLM := mock.NewMockLLM()
	mockDocker := docker.NewMockDriver()

	// Create new pipeline with mock implementations
	p := v2.NewPipeline(
		v2.WithLLM(mockLLM),
		v2.WithDockerDriver(mockDocker),
	)

	// Get absolute path to test project
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the pipeline
	if err := p.Run(ctx, absPath); err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(absPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Fatalf("Dockerfile was not created")
	}
}
