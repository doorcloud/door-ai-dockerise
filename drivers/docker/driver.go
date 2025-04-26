package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Driver provides Docker operations using the Docker CLI
type Driver struct{}

// New creates a new Docker driver
func New() *Driver {
	return &Driver{}
}

// Build builds a Docker image from a Dockerfile
func (d *Driver) Build(ctx context.Context, dir, dockerfile string) (string, error) {
	// Write Dockerfile to disk
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %v", err)
	}

	// Generate a unique image name
	imageName := fmt.Sprintf("test-%d", os.Getpid())

	// Run docker build
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, ".")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker build failed: %v\n%s", err, output)
	}

	// Clean up Dockerfile
	if err := os.Remove(dockerfilePath); err != nil {
		return "", fmt.Errorf("failed to clean up Dockerfile: %v", err)
	}

	return imageName, nil
}

// Run runs a Docker container
func (d *Driver) Run(ctx context.Context, imageID string, port int) error {
	// Run docker container
	cmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", port, port), imageID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %v\n%s", err, output)
	}

	return nil
}
