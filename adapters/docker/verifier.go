package docker

import (
	"context"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// VerifierAdapter wraps a Docker driver to implement core.Verifier
type VerifierAdapter struct {
	driver docker.Driver
}

// NewVerifierAdapter creates a new verifier adapter
func NewVerifierAdapter(driver docker.Driver) *VerifierAdapter {
	return &VerifierAdapter{driver: driver}
}

// Verify implements core.Verifier
func (v *VerifierAdapter) Verify(ctx context.Context, root string, generatedFile string) error {
	// Write the Dockerfile to a temporary location
	dockerfilePath := filepath.Join(root, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(generatedFile), 0644); err != nil {
		return err
	}

	// Build and verify the Dockerfile
	return v.driver.Build(ctx, dockerfilePath, docker.BuildOptions{
		Context:    root,
		Tags:       []string{"test-image"},
		Dockerfile: "Dockerfile",
	})
}
