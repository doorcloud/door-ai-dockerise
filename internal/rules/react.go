package rules

import (
	"io/fs"
	"strings"
)

// React implements the types.Detector interface for React projects
type React struct{}

func (r *React) Name() string {
	return "react"
}

func (r *React) Detect(fsys fs.FS) (bool, error) {
	// Check for package.json with react-scripts or vite
	pkgContent, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false, nil
	}

	// Check for react-scripts or vite in dependencies
	if !strings.Contains(string(pkgContent), "react-scripts") && !strings.Contains(string(pkgContent), "vite") {
		return false, nil
	}

	// Check for src/index.js or src/index.tsx
	if _, err := fs.Stat(fsys, "src/index.js"); err == nil {
		return true, nil
	}
	if _, err := fs.Stat(fsys, "src/index.tsx"); err == nil {
		return true, nil
	}

	// Check for vite.config.js or vite.config.ts
	if _, err := fs.Stat(fsys, "vite.config.js"); err == nil {
		return true, nil
	}
	if _, err := fs.Stat(fsys, "vite.config.ts"); err == nil {
		return true, nil
	}

	return false, nil
}

func init() {
	NewRegistry().Register(&React{})
}
