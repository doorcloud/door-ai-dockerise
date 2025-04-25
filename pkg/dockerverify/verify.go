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

// copyBuildContext copies files from fsys to the build context directory
func copyBuildContext(fsys fs.FS, dstDir string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if not needed in build context
		if !shouldCopy(path) {
			return nil
		}

		// Get file info and mode
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("get file info: %w", err)
		}

		// Create destination path
		dstPath := filepath.Join(dstDir, path)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("create parent dirs: %w", err)
		}

		// Copy directories
		if d.IsDir() {
			return os.Mkdir(dstPath, info.Mode())
		}

		// Copy files
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		return nil
	})
}
