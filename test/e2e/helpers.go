package e2e

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	dockermock "github.com/doorcloud/door-ai-dockerise/drivers/docker/mock"
	"github.com/doorcloud/door-ai-dockerise/pipeline"
)

// RunPipeline runs the pipeline on the given source directory
func RunPipeline(t *testing.T, sourceDir string) error {
	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create pipeline with mock components
	p := pipeline.New(pipeline.Options{
		Detectors: detectors.Registry(),
		FactProviders: []core.FactProvider{
			facts.NewStatic(),
		},
		Generator:  generate.NewLLM(mockLLM),
		Verifier:   dockermock.NewMockDriver(),
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

type mockFactProvider struct{}

func (m *mockFactProvider) Facts(ctx context.Context, stack core.StackInfo) ([]core.Fact, error) {
	return []core.Fact{
		{Key: "stack_type", Value: stack.Name},
		{Key: "build_tool", Value: stack.BuildTool},
	}, nil
}

func setupTestEnvironment(t *testing.T) (*pipeline.Pipeline, string) {
	// Create test directory
	testDir, err := os.MkdirTemp("", "dockerfile-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create Docker driver
	dockerDriver := dockermock.NewMockDriver()

	// Create pipeline with detectors
	p := pipeline.New(pipeline.Options{
		Detectors: detectors.Registry(),
		FactProviders: []core.FactProvider{
			&mockFactProvider{},
		},
		Generator: &mock.MockGenerator{},
		Verifier:  dockerDriver,
	})

	return p, testDir
}

// copyDir recursively copies a directory tree
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == src {
			return nil
		}

		// Create the destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// Create directories
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy files
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
