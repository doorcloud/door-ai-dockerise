package legacyshim

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Wrapper adapts a legacy detector to the new interface
type Wrapper struct {
	detector core.Detector
}

// NewWrapper creates a new wrapper around a legacy detector
func NewWrapper(detector core.Detector) *Wrapper {
	return &Wrapper{detector: detector}
}

// Detect implements the core.Detector interface
func (w *Wrapper) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "legacyshim-*")
	if err != nil {
		return core.StackInfo{}, err
	}
	defer os.RemoveAll(tmpDir)

	// Copy files from fs.FS to temporary directory
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Create directories
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(tmpDir, path), 0755)
		}

		// Copy files
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		return os.WriteFile(filepath.Join(tmpDir, path), data, 0644)
	})
	if err != nil {
		return core.StackInfo{}, err
	}

	// Call the legacy detector with the temporary directory
	return w.detector.Detect(ctx, os.DirFS(tmpDir))
}
