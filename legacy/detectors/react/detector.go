package react

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// ReactDetector detects React projects
type ReactDetector struct{}

// NewReactDetector creates a new ReactDetector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{}
}

// Detect checks if the given directory contains a React project
func (d *ReactDetector) Detect(ctx context.Context, dir string) (*Info, error) {
	// Check for package.json
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return &Info{}, nil
	}

	// Check for React dependencies
	packageJSON, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return &Info{}, err
	}

	// Simple check for React in package.json
	if !containsReact(string(packageJSON)) {
		return &Info{}, nil
	}

	return &Info{
		BuildTool: "npm",
	}, nil
}

// Info contains information about a detected React project
type Info struct {
	BuildTool string
}

// containsReact checks if the package.json contains React dependencies
func containsReact(packageJSON string) bool {
	// Simple check for React in package.json
	return strings.Contains(packageJSON, `"react"`) || strings.Contains(packageJSON, `"@types/react"`)
}
