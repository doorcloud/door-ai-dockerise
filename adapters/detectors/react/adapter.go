package react

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// ReactDetector implements the core.Detector interface for React projects
type ReactDetector struct{}

// NewReactDetector creates a new ReactDetector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{}
}

// Detect implements the core.Detector interface
func (d *ReactDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, error) {
	// Check for package.json
	packageJSON, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		if err == fs.ErrNotExist {
			return core.StackInfo{}, nil
		}
		return core.StackInfo{}, err
	}

	// Check for React dependencies
	if !containsReact(string(packageJSON)) {
		return core.StackInfo{}, nil
	}

	return core.StackInfo{
		Name:      "react",
		BuildTool: "npm",
	}, nil
}

// containsReact checks if the package.json contains React dependencies
func containsReact(packageJSON string) bool {
	return strings.Contains(packageJSON, `"react"`) || strings.Contains(packageJSON, `"@types/react"`)
}

// Name returns the detector name
func (d *ReactDetector) Name() string {
	return "react"
}
