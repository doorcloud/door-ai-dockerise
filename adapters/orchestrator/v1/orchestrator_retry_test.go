package v1

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/logs"
)

type mockBuilder struct {
	attempts int
	maxFails int
}

func (m *mockBuilder) Build(ctx context.Context, in core.BuildInput, log core.LogStreamer) (core.ImageRef, error) {
	m.attempts++
	if m.attempts <= m.maxFails {
		log.Error("Build failed")
		return core.ImageRef{}, fmt.Errorf("build failed (attempt %d)", m.attempts)
	}
	log.Info("Build succeeded")
	return core.ImageRef{Name: "test:latest"}, nil
}

type mockDetector struct {
	logSink core.LogSink
}

func (m *mockDetector) Detect(ctx context.Context, root fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	info := core.StackInfo{
		Name:      "test",
		BuildTool: "test",
		Version:   "1.0",
	}

	if logSink != nil {
		logs.Tag("detect", "detector=%s found=true path=%s", m.Name(), "test")
	}

	return info, true, nil
}

func (d *mockDetector) Name() string {
	return "mock"
}

func (d *mockDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}

type mockGenerator struct{}

func (m *mockGenerator) Generate(ctx context.Context, facts core.Facts) (string, error) {
	return "FROM alpine", nil
}

func (m *mockGenerator) Fix(ctx context.Context, dockerfile string, error string) (string, error) {
	return dockerfile, nil
}

type mockFactProvider struct{}

func (m *mockFactProvider) Facts(ctx context.Context, stack core.StackInfo) ([]core.Fact, error) {
	return []core.Fact{
		{Key: "stack_type", Value: stack.Name},
		{Key: "build_tool", Value: stack.BuildTool},
	}, nil
}

func TestOrchestrator_Retry(t *testing.T) {
	builder := &mockBuilder{maxFails: 2}

	o := New(Opts{
		Builder:   builder,
		Attempts:  3,
		Detectors: []core.Detector{&mockDetector{}},
		Generator: &mockGenerator{},
		Facts:     []core.FactProvider{&mockFactProvider{}},
	})

	// Run the orchestrator
	_, err := o.Run(context.Background(), ".", nil, io.Discard)
	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	// Check that we retried the correct number of times
	if builder.attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", builder.attempts)
	}
}
