package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/log/mock"
	"github.com/doorcloud/door-ai-dockerise/core"
)

func createTestTar(t *testing.T, dockerfile string) io.Reader {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	// Add Dockerfile to tar
	err := tw.WriteHeader(&tar.Header{
		Name: "Dockerfile",
		Mode: 0o644,
		Size: int64(len(dockerfile)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		t.Fatal(err)
	}

	return &buf
}

func TestEngine_Build(t *testing.T) {
	ctx := context.Background()
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	t.Run("successful build", func(t *testing.T) {
		// Create a temporary directory
		dir := t.TempDir()

		// Create a valid Dockerfile
		dockerfile := "FROM alpine:latest\nRUN echo 'test'"
		err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0o644)
		if err != nil {
			t.Fatalf("Failed to write Dockerfile: %v", err)
		}

		log := mock.New()
		input := core.BuildInput{
			ContextTar: createTestTar(t, dockerfile),
			Dockerfile: dockerfile,
		}

		_, err = engine.Build(ctx, input, log)
		if err != nil {
			t.Errorf("Build failed: %v", err)
		}

		entries := log.Entries()
		if len(entries) == 0 {
			t.Error("Expected log entries, got none")
		}
	})

	t.Run("failed build", func(t *testing.T) {
		// Create an invalid Dockerfile
		dockerfile := "INVALID FROM syntax\nBAD RUN command"
		log := mock.New()
		input := core.BuildInput{
			ContextTar: createTestTar(t, dockerfile),
			Dockerfile: dockerfile,
		}

		_, err := engine.Build(ctx, input, log)
		if err == nil {
			t.Error("Expected build to fail")
		}

		entries := log.Entries()
		if len(entries) == 0 {
			t.Error("Expected error log entries, got none")
		}
	})
}
