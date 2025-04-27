package docker

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Verifier verifies Dockerfiles by building and running them
type Verifier struct{}

// NewVerifier creates a new Docker verifier
func NewVerifier() *Verifier {
	return &Verifier{}
}

// Verify builds and runs a Dockerfile to verify it works
func (v *Verifier) Verify(ctx context.Context, dir string, dockerfile string, port int, logs io.Writer) error {
	// Build the image
	buildCmd := exec.CommandContext(ctx, "docker", "build", "-t", "test-image", "-f", dockerfile, ".")
	buildCmd.Dir = dir
	buildCmd.Stdout = logs
	buildCmd.Stderr = logs
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Run the container
	runCmd := exec.CommandContext(ctx, "docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", port, port), "test-image")
	runCmd.Dir = dir
	runCmd.Stdout = logs
	runCmd.Stderr = logs
	if err := runCmd.Run(); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	// TODO: Add health check to verify the application is running

	return nil
}
