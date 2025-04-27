package e2e

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

// RunPipeline runs the pipeline on the given source directory
func RunPipeline(t *testing.T, sourceDir string) error {
	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline with mock components
	p := v2.New(v2.Options{
		Detectors: []core.Detector{
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		},
		FactProviders: []core.FactProvider{
			facts.NewStatic(),
		},
		Generator:  generate.NewLLM(mockLLM),
		Verifier:   docker.NewMockDriver(),
		MaxRetries: 3,
	})

	// Create test context
	ctx := context.Background()

	// Get absolute path to source directory
	absPath, err := filepath.Abs(sourceDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the pipeline
	return p.Run(ctx, absPath)
}

// SetupTestDir creates a temporary directory for testing
func SetupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "dockerfile-gen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

type testLogger struct {
	t *testing.T
}

func (l *testLogger) Write(p []byte) (int, error) {
	l.t.Logf("%s", bytes.TrimSpace(p))
	return len(p), nil
}

func newTestVerifier(t *testing.T) *verifiers.Docker {
	verifier, err := verifiers.NewDocker(verifiers.Options{
		Socket:  "unix:///var/run/docker.sock",
		LogSink: os.Stdout,
		Timeout: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Failed to create Docker verifier: %v", err)
	}
	return verifier
}

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "dockerfile-gen-test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
