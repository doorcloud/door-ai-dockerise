package core

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core/logs"
	"github.com/stretchr/testify/assert"
)

type mockDetector struct {
	stack   StackInfo
	err     error
	logSink LogSink
}

func (d *mockDetector) Detect(ctx context.Context, fsys fs.FS, logSink LogSink) (StackInfo, bool, error) {
	if logSink != nil && d.stack.Name != "" {
		logs.Tag("detect", "detector=%s found=true path=%s", d.Name(), d.stack.DetectedFiles[0])
	}
	return d.stack, d.stack.Name != "", d.err
}

func (d *mockDetector) Name() string {
	return "mock"
}

func (d *mockDetector) Describe() string {
	return "Mock detector for testing"
}

func (d *mockDetector) SetLogSink(logSink LogSink) {
	d.logSink = logSink
}

type mockVerifier struct {
	err error
}

func (v *mockVerifier) Verify(ctx context.Context, root string, dockerfile string) error {
	return v.err
}

type mockGenerator struct {
	dockerfile string
	err        error
}

func (g *mockGenerator) Generate(ctx context.Context, facts Facts) (string, error) {
	return g.dockerfile, g.err
}

func (g *mockGenerator) Fix(ctx context.Context, dockerfile string, error string) (string, error) {
	return g.dockerfile, g.err
}

func newMockGenerator(dockerfile string, err error) *mockGenerator {
	return &mockGenerator{
		dockerfile: dockerfile,
		err:        err,
	}
}

func TestOrchestrator_Run_WithSpec(t *testing.T) {
	// Setup
	detector := &mockDetector{}
	generator := newMockGenerator("FROM node:14", nil)
	verifier := &mockVerifier{}

	o := NewOrchestrator(detector, generator, verifier)

	// Test
	spec := &Spec{
		Framework: "node",
		BuildTool: "npm",
		Version:   "14",
	}

	dockerfile, err := o.Run(context.Background(), ".", spec)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "FROM node:14", dockerfile)
}

func TestOrchestrator_Run_DetectError(t *testing.T) {
	// Setup
	detector := &mockDetector{err: errors.New("detection failed")}
	generator := newMockGenerator("", nil)
	verifier := &mockVerifier{}

	o := NewOrchestrator(detector, generator, verifier)

	// Test
	dockerfile, err := o.Run(context.Background(), ".", nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detection failed")
	assert.Empty(t, dockerfile)
}

func TestOrchestrator_Run_GenerateError(t *testing.T) {
	// Setup
	detector := &mockDetector{stack: StackInfo{Name: "node"}}
	generator := newMockGenerator("", errors.New("generation failed"))
	verifier := &mockVerifier{}

	o := NewOrchestrator(detector, generator, verifier)

	// Test
	dockerfile, err := o.Run(context.Background(), ".", nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generation failed")
	assert.Empty(t, dockerfile)
}

func TestOrchestrator_Run_VerifyError(t *testing.T) {
	// Setup
	detector := &mockDetector{stack: StackInfo{Name: "node"}}
	generator := newMockGenerator("FROM node:14", nil)
	verifier := &mockVerifier{err: errors.New("verification failed")}

	o := NewOrchestrator(detector, generator, verifier)

	// Test
	spec := &Spec{
		Framework: "node",
		BuildTool: "npm",
		Version:   "14",
	}

	dockerfile, err := o.Run(context.Background(), ".", spec)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")
	assert.Empty(t, dockerfile)
}

func TestOrchestrator_Run_Success(t *testing.T) {
	// Setup
	detector := &mockDetector{stack: StackInfo{Name: "node"}}
	generator := newMockGenerator("FROM node:14", nil)
	verifier := &mockVerifier{}

	o := NewOrchestrator(detector, generator, verifier)

	// Test
	spec := &Spec{
		Framework: "node",
		BuildTool: "npm",
		Version:   "14",
	}

	dockerfile, err := o.Run(context.Background(), ".", spec)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "FROM node:14", dockerfile)
}
