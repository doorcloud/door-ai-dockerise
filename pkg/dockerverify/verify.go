package dockerverify

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Verify checks if the Dockerfile is valid and can be built
func Verify(ctx context.Context, fsys fs.FS, dockerfile string) error {
	// Strip markdown fences if present
	dockerfile = strings.TrimSpace(dockerfile)
	dockerfile = strings.TrimPrefix(dockerfile, "```dockerfile")
	dockerfile = strings.TrimPrefix(dockerfile, "```")
	dockerfile = strings.TrimSuffix(dockerfile, "```")

	// Create temporary directory for build context
	tmpDir, err := os.MkdirTemp("", "dockerfile-verify-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write Dockerfile
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return fmt.Errorf("write dockerfile: %w", err)
	}

	// Copy build context
	if err := copyBuildContext(fsys, tmpDir); err != nil {
		return fmt.Errorf("copy build context: %w", err)
	}

	// TODO: Build the Dockerfile
	return nil
}
