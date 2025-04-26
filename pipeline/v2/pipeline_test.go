package v2

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
)

func TestPipeline_Run(t *testing.T) {
	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline with mock components
	p := NewPipeline(
		WithDetectors(
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		WithFactProviders(
			facts.NewStatic(),
		),
		WithGenerator(generate.NewLLM(mockLLM)),
		WithDockerDriver(docker.NewMockDriver()),
		WithMaxRetries(3),
	)

	// Create test context
	ctx := context.Background()

	// Get absolute paths to test projects
	reactPath, err := filepath.Abs("testdata/react-project")
	if err != nil {
		t.Fatalf("Failed to get absolute path for React project: %v", err)
	}

	springbootPath, err := filepath.Abs("testdata/springboot-project")
	if err != nil {
		t.Fatalf("Failed to get absolute path for Spring Boot project: %v", err)
	}

	// Test with a React project
	err = p.Run(ctx, reactPath)
	if err != nil {
		t.Errorf("Pipeline.Run() error with React project = %v", err)
	}

	// Test with a Spring Boot project
	err = p.Run(ctx, springbootPath)
	if err != nil {
		t.Errorf("Pipeline.Run() error with Spring Boot project = %v", err)
	}
}

func TestPipeline_Run_ErrorCases(t *testing.T) {
	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline with mock components
	p := NewPipeline(
		WithDetectors(
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		WithFactProviders(
			facts.NewStatic(),
		),
		WithGenerator(generate.NewLLM(mockLLM)),
		WithDockerDriver(docker.NewMockDriver()),
		WithMaxRetries(3),
	)

	// Create test context
	ctx := context.Background()

	// Test with non-existent directory
	err := p.Run(ctx, "non-existent-directory")
	if err == nil {
		t.Error("Pipeline.Run() expected error for non-existent directory")
	}

	// Create empty directory
	emptyDir, err := os.MkdirTemp("", "empty-project-*")
	if err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	// Test with empty directory
	err = p.Run(ctx, emptyDir)
	if err == nil {
		t.Error("Pipeline.Run() expected error for empty directory")
	}
}
