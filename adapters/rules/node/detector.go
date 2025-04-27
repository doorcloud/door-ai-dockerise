package node

import (
	"encoding/json"
	"io/fs"
)

// NodeDetector implements detection rules for Node.js projects
type NodeDetector struct{}

// Detect checks if the given filesystem contains a Node.js project
func (d NodeDetector) Detect(fsys fs.FS) bool {
	// Check for package.json
	packageJSON, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false
	}

	// Parse package.json
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Main            string            `json:"main"`
	}
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		return false
	}

	// Check for React dependencies
	if _, hasReact := pkg.Dependencies["react"]; hasReact {
		return false
	}

	// Check for Node.js-specific dependencies
	hasNodeDeps := false
	for dep := range pkg.Dependencies {
		switch dep {
		case "express", "koa", "fastify", "hapi", "nest", "socket.io":
			hasNodeDeps = true
		}
	}

	// Check for main entry point
	if pkg.Main != "" {
		if _, err := fs.ReadFile(fsys, pkg.Main); err == nil {
			return true
		}
	}

	// Check for common Node.js files
	nodeFiles := []string{"index.js", "server.js", "app.js"}
	for _, file := range nodeFiles {
		if _, err := fs.ReadFile(fsys, file); err == nil {
			return true
		}
	}

	return hasNodeDeps
}
