package verifiers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Docker struct{}

func NewDocker() *Docker {
	return &Docker{}
}

func (d *Docker) Verify(ctx context.Context, repoPath, dockerfile string) error {
	// Build the image
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", "test-image", "-f", "-", repoPath)
	cmd.Stdin = strings.NewReader(dockerfile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build image: %v\nOutput: %s", err, output)
	}

	// Run the container
	cmd = exec.CommandContext(ctx, "docker", "run", "--rm", "test-image")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run container: %v\nOutput: %s", err, output)
	}

	return nil
}
