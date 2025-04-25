package rules

import (
	"io/fs"
)

// nodeDetector implements the types.Detector interface
type nodeDetector struct{}

// Detect checks if the project is a Node.js project
func (d *nodeDetector) Detect(fsys fs.FS) (bool, error) {
	// Check for package.json
	_, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false, nil
	}

	// Check for node_modules directory
	_, err = fs.Stat(fsys, "node_modules")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Name returns the name of the detector
func (d *nodeDetector) Name() string {
	return "node"
}
