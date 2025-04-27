package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/doorcloud/door-ai-dockerise/core/errs"
)

// Driver defines the interface for Docker operations
type Driver interface {
	// Verify checks if the Dockerfile is valid
	Verify(ctx context.Context, dockerfile string) error
	// Build builds a Docker image from a Dockerfile
	Build(ctx context.Context, dockerfilePath string, opts BuildOptions) error
	// Push pushes a Docker image to a registry
	Push(ctx context.Context, image string) error
	// BuildDockerfile builds a Docker image from a Dockerfile
	BuildDockerfile(ctx context.Context, dir, dockerfile string) (string, error)
	// Run runs a Docker container and returns its ID
	Run(ctx context.Context, imageID string, hostPort, containerPort int) (string, error)
	// GetRandomPort returns a random available port
	GetRandomPort() (int, error)
	// Stop stops a Docker container
	Stop(ctx context.Context, containerID string) error
	// Remove removes a Docker container
	Remove(ctx context.Context, containerID string) error
	// Logs streams container logs
	Logs(ctx context.Context, containerID string, w io.Writer) error
}

// dockerDriver implements the Driver interface
type dockerDriver struct {
	timeout time.Duration
}

// New creates a new Docker driver
func New() Driver {
	return &dockerDriver{
		timeout: 5 * time.Minute,
	}
}

// Verify checks if the Dockerfile is valid
func (d *dockerDriver) Verify(ctx context.Context, dockerfile string) error {
	// Create a temporary file with the Dockerfile content
	tmpFile, err := os.CreateTemp("", "dockerfile-*.dockerfile")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(dockerfile); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Run docker build --dry-run
	cmd := exec.CommandContext(ctx, "docker", "build", "--dry-run", "-f", tmpFile.Name(), ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Dockerfile validation failed: %w", err)
	}

	return nil
}

// Build builds a Docker image from a Dockerfile
func (d *dockerDriver) Build(ctx context.Context, dockerfilePath string, opts BuildOptions) error {
	// Prepare docker build command
	args := []string{
		"build",
		"-f", dockerfilePath,
	}

	// Add tags
	for _, tag := range opts.Tags {
		args = append(args, "-t", tag)
	}

	// Add context
	args = append(args, opts.Context)

	// Run docker build
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// Push pushes a Docker image to a registry
func (d *dockerDriver) Push(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "push", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errs.Wrap("docker push", fmt.Errorf("%v\nOutput: %s", err, output))
	}
	return nil
}

// BuildDockerfile builds a Docker image from a Dockerfile
func (d *dockerDriver) BuildDockerfile(ctx context.Context, dir, dockerfile string) (string, error) {
	// Write Dockerfile to disk
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Build the image
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", "temp-image", dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w\nOutput: %s", err, output)
	}

	return "temp-image", nil
}

// Run runs a Docker container and returns its ID
func (d *dockerDriver) Run(ctx context.Context, imageID string, hostPort, containerPort int) (string, error) {
	// Run docker container
	cmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", hostPort, containerPort), imageID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errs.Wrap("docker run", fmt.Errorf("%v\n%s", err, output))
	}

	// Return container ID (trim newline)
	return strings.TrimSpace(string(output)), nil
}

// GetRandomPort returns a random available port
func (d *dockerDriver) GetRandomPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to get random port: %w", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// Stop stops a Docker container
func (d *dockerDriver) Stop(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errs.Wrap("docker stop", fmt.Errorf("%v\n%s", err, output))
	}
	return nil
}

// Remove removes a Docker container
func (d *dockerDriver) Remove(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errs.Wrap("docker rm", fmt.Errorf("%v\n%s", err, output))
	}
	return nil
}

// Logs streams container logs
func (d *dockerDriver) Logs(ctx context.Context, containerID string, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "docker", "logs", "-f", containerID)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}
