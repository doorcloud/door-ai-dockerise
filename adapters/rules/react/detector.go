package react

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// ReactDetector implements detection rules for React projects
type ReactDetector struct {
	logSink core.LogSink
}

// NewReactDetector creates a new React detector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{}
}

// Name returns the detector name
func (d *ReactDetector) Name() string {
	return "react"
}

// Describe returns a description of what the detector looks for
func (d *ReactDetector) Describe() string {
	return "Detects React projects by checking for React dependencies and React-specific files"
}

// SetLogSink sets the log sink for the detector
func (d *ReactDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}

// Detect checks if the given filesystem contains a React project
func (d *ReactDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	// Check for package.json
	pkgJson, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return core.StackInfo{}, false, nil
	}

	// Check for React in package.json
	if !strings.Contains(string(pkgJson), `"react"`) {
		return core.StackInfo{}, false, nil
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

	if !hasReactFiles && !strings.Contains(string(pkgJson), `"react-scripts"`) {
		return core.StackInfo{}, false, nil
	}

	if d.logSink != nil {
		d.logSink.Log("detector=react found=true")
	}

	return core.StackInfo{
		Name:      "react",
		BuildTool: "npm",
		DetectedFiles: []string{
			"package.json",
			"src/index.js",
		},
	}, true, nil
}
