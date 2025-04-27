package mock

import (
	"context"
	"io"

	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// MockDriver implements the Docker driver interface for testing
type MockDriver struct {
	GetRandomPortFunc   func() (int, error)
	RunFunc             func(ctx context.Context, image string, hostPort, containerPort int) (string, error)
	LogsFunc            func(ctx context.Context, containerID string, w io.Writer) error
	StopFunc            func(ctx context.Context, containerID string) error
	RemoveFunc          func(ctx context.Context, containerID string) error
	BuildFunc           func(ctx context.Context, dockerfilePath string, opts docker.BuildOptions) error
	BuildDockerfileFunc func(ctx context.Context, dir, dockerfile string) (string, error)
	PushFunc            func(ctx context.Context, image string) error
	VerifyFunc          func(ctx context.Context, dockerfile string) error
}

// NewMockDriver creates a new mock Docker driver
func NewMockDriver() *MockDriver {
	return &MockDriver{
		GetRandomPortFunc: func() (int, error) {
			return 8080, nil
		},
		RunFunc: func(ctx context.Context, image string, hostPort, containerPort int) (string, error) {
			return "test-container-id", nil
		},
		LogsFunc: func(ctx context.Context, containerID string, w io.Writer) error {
			return nil
		},
		StopFunc: func(ctx context.Context, containerID string) error {
			return nil
		},
		RemoveFunc: func(ctx context.Context, containerID string) error {
			return nil
		},
		BuildFunc: func(ctx context.Context, dockerfilePath string, opts docker.BuildOptions) error {
			return nil
		},
		BuildDockerfileFunc: func(ctx context.Context, dir, dockerfile string) (string, error) {
			return "test-image-id", nil
		},
		PushFunc: func(ctx context.Context, image string) error {
			return nil
		},
		VerifyFunc: func(ctx context.Context, dockerfile string) error {
			return nil
		},
	}
}

// GetRandomPort implements the Docker driver interface
func (m *MockDriver) GetRandomPort() (int, error) {
	return m.GetRandomPortFunc()
}

// Run implements the Docker driver interface
func (m *MockDriver) Run(ctx context.Context, image string, hostPort, containerPort int) (string, error) {
	return m.RunFunc(ctx, image, hostPort, containerPort)
}

// Logs implements the Docker driver interface
func (m *MockDriver) Logs(ctx context.Context, containerID string, w io.Writer) error {
	return m.LogsFunc(ctx, containerID, w)
}

// Stop implements the Docker driver interface
func (m *MockDriver) Stop(ctx context.Context, containerID string) error {
	return m.StopFunc(ctx, containerID)
}

// Remove implements the Docker driver interface
func (m *MockDriver) Remove(ctx context.Context, containerID string) error {
	return m.RemoveFunc(ctx, containerID)
}

// Build implements the Docker driver interface
func (m *MockDriver) Build(ctx context.Context, dockerfilePath string, opts docker.BuildOptions) error {
	return m.BuildFunc(ctx, dockerfilePath, opts)
}

// BuildDockerfile implements the Docker driver interface
func (m *MockDriver) BuildDockerfile(ctx context.Context, dir, dockerfile string) (string, error) {
	return m.BuildDockerfileFunc(ctx, dir, dockerfile)
}

// Push implements the Docker driver interface
func (m *MockDriver) Push(ctx context.Context, image string) error {
	return m.PushFunc(ctx, image)
}

// Verify implements the Docker driver interface
func (m *MockDriver) Verify(ctx context.Context, dockerfile string) error {
	return m.VerifyFunc(ctx, dockerfile)
}
