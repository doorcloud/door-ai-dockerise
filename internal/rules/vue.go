package rules

import (
	"io/fs"
	"strings"
)

// vueDetector implements the types.Detector interface
type vueDetector struct{}

// Detect checks if the project is a Vue.js project
func (d *vueDetector) Detect(fsys fs.FS) (bool, error) {
	// Check for package.json
	content, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false, nil
	}

	// Check for vue dependency
	if !strings.Contains(string(content), `"vue"`) {
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
func (d *vueDetector) Name() string {
	return "vue"
}
