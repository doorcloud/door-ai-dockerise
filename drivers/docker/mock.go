package docker

import (
	"context"
)

// MockDriver is a mock implementation of the Driver interface
type MockDriver struct {
	verifyErr error
	buildErr  error
}

// NewMockDriver creates a new mock Docker driver
func NewMockDriver() *MockDriver {
	return &MockDriver{}
}

// WithVerifyError sets the error to return from Verify
func (d *MockDriver) WithVerifyError(err error) *MockDriver {
	d.verifyErr = err
	return d
}

// WithBuildError sets the error to return from Build
func (d *MockDriver) WithBuildError(err error) *MockDriver {
	d.buildErr = err
	return d
}

// Verify implements the Driver interface
func (d *MockDriver) Verify(ctx context.Context, dockerfile string) error {
	return d.verifyErr
}

// Build implements the Driver interface
func (d *MockDriver) Build(ctx context.Context, dockerfilePath string, opts BuildOptions) error {
	return d.buildErr
}

// Push implements the Docker driver interface
func (d *MockDriver) Push(ctx context.Context, image string) error {
	return nil
}
