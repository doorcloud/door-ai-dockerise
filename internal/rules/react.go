package rules

import (
	"io/fs"
	"strings"
)

// reactDetector implements the types.Detector interface
type reactDetector struct{}

// Detect checks if the project is a React project
func (d *reactDetector) Detect(fsys fs.FS) (bool, error) {
	// Check for package.json
	content, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false, nil
	}

	// Check for react dependency
	if !strings.Contains(string(content), `"react"`) {
		return false, nil
	}

	// Check for src directory
	_, err = fs.Stat(fsys, "src")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Name returns the name of the detector
func (d *reactDetector) Name() string {
	return "react"
}
