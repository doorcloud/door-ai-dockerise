package core

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDetector implements the Detector interface for testing
type MockDetector struct {
	called bool
	err    error
}

func (m *MockDetector) Detect(ctx context.Context, fsys fs.FS) (StackInfo, error) {
	m.called = true
	if m.err != nil {
		return StackInfo{}, m.err
	}
	return StackInfo{
		Name:      "test-stack",
		BuildTool: "test-tool",
		Version:   "1.0.0",
	}, nil
}

// MockGenerator implements the Generator interface for testing
type MockGenerator struct {
	called bool
	err    error
}

func (m *MockGenerator) Generate(ctx context.Context, stack StackInfo, facts []Fact) (string, error) {
	m.called = true
	if m.err != nil {
		return "", m.err
	}
	return "FROM test-stack:1.0.0", nil
}

// MockVerifier implements the Verifier interface for testing
type MockVerifier struct {
	called bool
	err    error
}

func (m *MockVerifier) Verify(ctx context.Context, root string, generatedFile string) error {
	m.called = true
	return m.err
}

func TestOrchestrator_WithSpec(t *testing.T) {
	// Create mock components
	detector := &MockDetector{}
	generator := &MockGenerator{}
	verifier := &MockVerifier{}

	// Create orchestrator with mock components
	o := NewOrchestrator(detector, generator, verifier)

	// Create a test spec
	spec := &Spec{
		Language:  "javascript",
		Framework: "node",
		Version:   "18",
		BuildTool: "npm",
		Params:    map[string]string{},
	}

	// Run the orchestrator with the spec
	_, err := o.Run(context.Background(), "/test/path", spec, nil)
	require.NoError(t, err)

	// Verify that the detector was not called
	assert.False(t, detector.called, "Detector should not be called when spec is provided")

	// Verify that generator and verifier were called
	assert.True(t, generator.called, "Generator should be called")
	assert.True(t, verifier.called, "Verifier should be called")
}

func TestOrchestrator_WithoutSpec(t *testing.T) {
	// Create mock components
	detector := &MockDetector{}
	generator := &MockGenerator{}
	verifier := &MockVerifier{}

	// Create orchestrator with mock components
	o := NewOrchestrator(detector, generator, verifier)

	// Run the orchestrator without a spec
	_, err := o.Run(context.Background(), "/test/path", nil, nil)
	require.NoError(t, err)

	// Verify that all components were called
	assert.True(t, detector.called, "Detector should be called when no spec is provided")
	assert.True(t, generator.called, "Generator should be called")
	assert.True(t, verifier.called, "Verifier should be called")
}

func TestOrchestrator_DetectionError(t *testing.T) {
	// Create mock components with detector error
	detector := &MockDetector{err: errors.New("detection failed")}
	generator := &MockGenerator{}
	verifier := &MockVerifier{}

	// Create orchestrator with mock components
	o := NewOrchestrator(detector, generator, verifier)

	// Run the orchestrator without a spec
	_, err := o.Run(context.Background(), "/test/path", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detection failed")

	// Verify that only detector was called
	assert.True(t, detector.called)
	assert.False(t, generator.called)
	assert.False(t, verifier.called)
}

func TestOrchestrator_GenerationError(t *testing.T) {
	// Create mock components with generator error
	detector := &MockDetector{}
	generator := &MockGenerator{err: errors.New("generation failed")}
	verifier := &MockVerifier{}

	// Create orchestrator with mock components
	o := NewOrchestrator(detector, generator, verifier)

	// Run the orchestrator with a spec
	_, err := o.Run(context.Background(), "/test/path", &Spec{
		Language:  "javascript",
		Framework: "node",
	}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generation failed")

	// Verify that detector was not called but generator was
	assert.False(t, detector.called)
	assert.True(t, generator.called)
	assert.False(t, verifier.called)
}

func TestOrchestrator_VerificationError(t *testing.T) {
	// Create mock components with verifier error
	detector := &MockDetector{}
	generator := &MockGenerator{}
	verifier := &MockVerifier{err: errors.New("verification failed")}

	// Create orchestrator with mock components
	o := NewOrchestrator(detector, generator, verifier)

	// Run the orchestrator with a spec
	_, err := o.Run(context.Background(), "/test/path", &Spec{
		Language:  "javascript",
		Framework: "node",
	}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")

	// Verify that all components were called
	assert.False(t, detector.called)
	assert.True(t, generator.called)
	assert.True(t, verifier.called)
}
