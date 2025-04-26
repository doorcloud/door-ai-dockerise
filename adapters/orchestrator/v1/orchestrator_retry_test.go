package v1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/mock"
	"github.com/stretchr/testify/assert"
)

type mockVerifier struct {
	failCount int
}

func (m *mockVerifier) Verify(ctx context.Context, root string, dockerfile string) error {
	if m.failCount > 0 {
		m.failCount--
		return fmt.Errorf("build failed")
	}
	return nil
}

func TestOrchestrator_Retry(t *testing.T) {
	// Create a mock generator that records calls
	gen := mock.NewMockClient()
	gen.SetResponse("test:test", "FROM ubuntu:latest\n")

	// Create a verifier that fails once
	verifier := &mockVerifier{failCount: 1}

	// Create orchestrator with 2 attempts and 10 minute timeout
	o := New(Opts{
		Detectors:    []core.Detector{},
		Facts:        []core.FactProvider{},
		Generator:    gen,
		Verifier:     verifier,
		Attempts:     2,
		BuildTimeout: 10,
	})

	// Run with a spec to avoid detection
	spec := &core.Spec{
		Framework: "test",
		BuildTool: "test",
	}

	// Test successful retry
	dockerfile, err := o.Run(context.Background(), ".", spec, nil)
	assert.NoError(t, err)
	assert.Contains(t, dockerfile, "FROM ubuntu:latest")

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
