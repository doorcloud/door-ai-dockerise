package dockerverify

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// SimpleVerifier implements Verifier using docker build
type SimpleVerifier struct{}

// New creates a new SimpleVerifier
func New() *SimpleVerifier {
	return &SimpleVerifier{}
}

// Verify attempts to build the Dockerfile in the given repository
func (v *SimpleVerifier) Verify(ctx context.Context, repo, dockerfile string, timeout time.Duration) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Write Dockerfile to repo
	if err := os.WriteFile(filepath.Join(repo, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return err
	}

	// Run docker build
	cmd := exec.CommandContext(ctx, "docker", "build", ".")
	cmd.Dir = repo
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
