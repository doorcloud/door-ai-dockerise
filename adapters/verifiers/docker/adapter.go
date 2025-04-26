package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DockerVerifier implements core.Verifier using Docker CLI
type DockerVerifier struct{}

// NewDockerVerifier creates a new DockerVerifier
func NewDockerVerifier() *DockerVerifier {
	return &DockerVerifier{}
}

// Verify implements the core.Verifier interface
func (v *DockerVerifier) Verify(ctx context.Context, root string, dockerfile string) error {
	// Write Dockerfile to disk
	dockerfilePath := filepath.Join(root, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return err
	}

	// Run docker build
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", "test", ".")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v\n%s", err, output)
	}

	// Clean up
	if err := os.Remove(dockerfilePath); err != nil {
		return fmt.Errorf("failed to clean up Dockerfile: %v", err)
	}

	return nil
}
