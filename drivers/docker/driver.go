package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// dockerDriver implements the Driver interface
type dockerDriver struct{}

// NewDriver creates a new Docker driver
func NewDriver() Driver {
	return &dockerDriver{}
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
		return fmt.Errorf("docker push failed: %v\nOutput: %s", err, output)
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

// Run runs a Docker container
func (d *dockerDriver) Run(ctx context.Context, imageID string, port int) error {
	// Run docker container
	cmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", port, port), imageID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %v\n%s", err, output)
	}

	return nil
}
