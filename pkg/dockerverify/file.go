package dockerverify

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CopyBuildContext(fsys fs.FS, dir string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip Dockerfile and .dockerignore
		if path == "Dockerfile" || path == ".dockerignore" {
			return nil
		}

		// Create directories
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dir, path), 0755)
		}

		// Determine if file should be copied
		shouldCopy := false

		// Always copy go.mod and go.sum
		if strings.HasSuffix(path, "go.mod") || strings.HasSuffix(path, "go.sum") {
			shouldCopy = true
		}

		// Copy Maven wrapper files
		if strings.HasSuffix(path, "mvnw") ||
			strings.HasSuffix(path, "mvnw.cmd") ||
			strings.Contains(path, ".mvn/wrapper/") {
			shouldCopy = true
		}

		if shouldCopy {
			// Create parent directories if needed
			if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0755); err != nil {
				return fmt.Errorf("create parent dirs for %s: %w", path, err)
			}

			// Copy file
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				return fmt.Errorf("read file %s: %w", path, err)
			}

			if err := os.WriteFile(filepath.Join(dir, path), data, 0644); err != nil {
				return fmt.Errorf("write file %s: %w", path, err)
			}
		}

		return nil
	})
}
