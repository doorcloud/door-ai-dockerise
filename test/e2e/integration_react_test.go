package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

func TestReactProject(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping integration test; set DG_E2E=1 to run")
	}

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline with mock components
	p := v2.NewPipeline(
		v2.WithDetectors(
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(docker.NewMockDriver()),
		v2.WithMaxRetries(3),
	)

	// Create test context
	ctx := context.Background()

	// Get absolute path to test project
	projectPath, err := filepath.Abs("testdata/react-project")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the pipeline
	if err := p.Run(ctx, projectPath); err != nil {
		t.Errorf("Pipeline.Run() error = %v", err)
	}

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Errorf("Dockerfile was not created at %s", dockerfilePath)
	}
}

func TestReactIntegration(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping integration test; set DG_E2E=1 to run")
	}

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline with mock components
	p := v2.NewPipeline(
		v2.WithDetectors(
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(docker.NewMockDriver()),
		v2.WithMaxRetries(3),
	)

	// Create test context
	ctx := context.Background()

	// Get absolute path to test project
	projectPath, err := filepath.Abs("testdata/react-project")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the pipeline
	if err := p.Run(ctx, projectPath); err != nil {
		t.Errorf("Pipeline.Run() error = %v", err)
	}

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Errorf("Dockerfile was not created at %s", dockerfilePath)
	}
}
