package e2e

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_Spring(t *testing.T) {
	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline with mock components
	p := pipeline.NewPipeline(
		pipeline.WithDetectors(
			springboot.NewSpringBootDetector(),
		),
		pipeline.WithFactProviders(
			facts.NewStatic(),
		),
		pipeline.WithGenerator(generate.NewLLM(mockLLM)),
		pipeline.WithDockerDriver(docker.NewMockDriver()),
		pipeline.WithMaxRetries(3),
		pipeline.WithLogSink(&logBuf),
	)

	// Create test context
	ctx := context.Background()

	// Create temporary directory
	dir, err := os.MkdirTemp("", "TestIntegration_Spring*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Copy test project to temp dir
	if err := copyDir("testdata/springboot-project", dir); err != nil {
		t.Fatalf("Failed to copy test project: %v", err)
	}

	// Run the pipeline
	err = p.Run(ctx, dir)
	if err != nil {
		t.Errorf("Pipeline.Run() error = %v", err)
	}

	// Check that Dockerfile was created
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Errorf("Dockerfile was not created at %s", dockerfilePath)
	}

	// Check log output
	logOutput := logBuf.String()
	t.Logf("Log output: %s", logOutput)
	assert.True(t, strings.Contains(logOutput, "detector=springboot found=true"), "Expected Spring Boot detector log line")
}
