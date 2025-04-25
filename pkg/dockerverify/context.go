package dockerverify

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

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

		// Make mvnw executable if present
		if filepath.Base(path) == "mvnw" {
			if err := os.Chmod(dstPath, 0755); err != nil {
				return fmt.Errorf("chmod mvnw: %w", err)
			}
		}

		return nil
	})
}
