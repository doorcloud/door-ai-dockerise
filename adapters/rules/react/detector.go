package react

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// ReactDetector implements detection rules for React projects
type ReactDetector struct{}

// Detect checks if the given filesystem contains a React project
func (d ReactDetector) Detect(fsys fs.FS) bool {
	// Check for package.json
	pkgJson, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false
	}

	// Check for React in package.json
	if !strings.Contains(string(pkgJson), `"react"`) {
		return false
	}

	// Check for React-specific files
	hasReactFiles := false
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() && (filepath.Ext(path) == ".jsx" || filepath.Ext(path) == ".tsx" || filepath.Ext(path) == ".js") {
			content, err := fs.ReadFile(fsys, path)
			if err == nil && (strings.Contains(string(content), "import React") || strings.Contains(string(content), "react-scripts")) {
				hasReactFiles = true
				return fs.SkipDir
			}
		}
		return nil
	})

	return hasReactFiles || strings.Contains(string(pkgJson), `"react-scripts"`)
}
