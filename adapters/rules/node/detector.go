package node

import (
	"io/fs"
)

// NodeDetector implements detection rules for Node.js projects
type NodeDetector struct{}

// Detect checks if the given filesystem contains a Node.js project
func (d NodeDetector) Detect(fsys fs.FS) bool {
	// Check for package.json
	_, err := fs.ReadFile(fsys, "package.json")
	return err == nil
}
