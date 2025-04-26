package docker

import (
	"context"
	"io"
)

// MockDriver is a mock implementation of the Docker driver
type MockDriver struct{}

// NewMockDriver creates a new mock Docker driver
func NewMockDriver() *MockDriver {
	return &MockDriver{}
}

// Build implements the Docker driver interface
func (d *MockDriver) Build(ctx context.Context, context io.Reader, options BuildOptions) error {
	return nil
}

// Push implements the Docker driver interface
func (d *MockDriver) Push(ctx context.Context, image string) error {
	return nil
}
