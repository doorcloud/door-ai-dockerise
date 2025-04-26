package docker

import (
	"context"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/log/mock"
	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestEngine_Build(t *testing.T) {
	ctx := context.Background()
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	t.Run("successful build", func(t *testing.T) {
		log := mock.New()
		input := core.BuildInput{
			ContextTar: nil, // TODO: Add test tar
			Dockerfile: "FROM alpine\nRUN echo 'test'",
		}

		_, err := engine.Build(ctx, input, log)
		if err != nil {
			t.Errorf("Build failed: %v", err)
		}

		entries := log.Entries()
		if len(entries) == 0 {
			t.Error("Expected log entries, got none")
		}
	})

	t.Run("failed build", func(t *testing.T) {
		log := mock.New()
		input := core.BuildInput{
			ContextTar: nil,
			Dockerfile: "INVALID DOCKERFILE",
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
