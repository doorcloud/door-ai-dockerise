package e2e

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/adapters/rules/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	coremock "github.com/doorcloud/door-ai-dockerise/core/mock"
	dockermock "github.com/doorcloud/door-ai-dockerise/drivers/docker/mock"
	"github.com/doorcloud/door-ai-dockerise/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestReactE2E(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test; set DG_E2E=1 to run")
	}

	// Create mock LLM
	mockLLM := coremock.NewMockLLM()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline with mock components
	p := pipeline.NewPipeline(
		pipeline.WithDetectors(
			react.NewReactDetector(),
			springboot.NewSpringBootDetector(),
		),
		pipeline.WithFactProviders(
			facts.NewStatic(),
		),
		pipeline.WithGenerator(generate.NewLLM(mockLLM)),
		pipeline.WithDockerDriver(dockermock.NewMockDriver()),
		pipeline.WithMaxRetries(3),
		pipeline.WithLogSink(&logBuf),
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

	// Verify log output
	logOutput := logBuf.String()
	assert.True(t, strings.Contains(logOutput, "detector=react found=true"), "Expected React detector log line")
	assert.False(t, strings.Contains(logOutput, "detector=spring-boot found=true"), "Unexpected Spring Boot detector log line")
}
