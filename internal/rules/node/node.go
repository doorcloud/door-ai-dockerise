package node

import (
	"io/fs"
)

// Detector implements types.Rule.
type Detector struct{}

func (Detector) Name() string {
	return "node"
}

func (Detector) Detect(fsys fs.FS) bool {
	// Check for package.json
	if _, err := fs.Stat(fsys, "package.json"); err != nil {
		return false
	}

	// Check for lock file to determine package manager
	if _, err := fs.Stat(fsys, "pnpm-lock.yaml"); err == nil {
		return true
	}
	if _, err := fs.Stat(fsys, "yarn.lock"); err == nil {
		return true
	}
	if _, err := fs.Stat(fsys, "package-lock.json"); err == nil {
		return true
	}

	return false
}

func (Detector) Facts(fsys fs.FS) map[string]any {
	// Determine package manager
	tool := "npm"
	if _, err := fs.Stat(fsys, "pnpm-lock.yaml"); err == nil {
		tool = "pnpm"
	} else if _, err := fs.Stat(fsys, "yarn.lock"); err == nil {
		tool = "yarn"
	}

	return map[string]any{
		"language":   "JavaScript",
		"framework":  "Node.js",
		"build_tool": tool,
		"build_cmd":  tool + " install",
		"start_cmd":  tool + " start",
		"ports":      []int{3000},
	}
}
