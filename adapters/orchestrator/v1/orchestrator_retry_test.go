package v1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/mock"
	"github.com/stretchr/testify/assert"
)

type mockBuilder struct {
	failCount int
	calls     int
}

func (m *mockBuilder) Build(ctx context.Context, in core.BuildInput, w io.Writer) (core.ImageRef, error) {
	m.calls++
	if m.failCount > 0 {
		m.failCount--
		return core.ImageRef{}, fmt.Errorf("build failed")
	}
	_, err := w.Write([]byte("built\n"))
	if err != nil {
		return core.ImageRef{}, err
	}
	return core.ImageRef{Name: "test-image:latest"}, nil
}

func TestOrchestrator_Retry(t *testing.T) {
	// Create a mock generator that records calls
	gen := mock.NewMockClient()
	gen.SetResponse("test:test", "FROM ubuntu:latest\n")

	// Create a builder that fails once
	builder := &mockBuilder{failCount: 1}

	// Create orchestrator with 2 attempts and 10 minute timeout
	o := New(Opts{
		Detectors:    []core.Detector{},
		Facts:        []core.FactProvider{},
		Generator:    gen,
		Builder:      builder,
		Attempts:     2,
		BuildTimeout: 10,
	})

	// Run with a spec to avoid detection
	spec := &core.Spec{
		Framework: "test",
		BuildTool: "test",
	}

	// Test successful retry
	var buf bytes.Buffer
	dockerfile, err := o.Run(context.Background(), ".", spec, &buf)
	assert.NoError(t, err)
	assert.Contains(t, dockerfile, "FROM ubuntu:latest")
	assert.Equal(t, 2, builder.calls) // Should have called Build twice
	assert.Contains(t, buf.String(), "built")

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = o.Run(ctx, ".", spec, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// Test build timeout
	timeoutCtx, _ := context.WithTimeout(context.Background(), 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond) // Ensure timeout is reached
	_, err = o.Run(timeoutCtx, ".", spec, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
