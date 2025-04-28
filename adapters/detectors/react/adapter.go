package react

import (
	"context"
	"encoding/json"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/logs"
)

// ReactDetector implements the core.Detector interface for React projects
type ReactDetector struct {
	logSink core.LogSink
}

// NewReactDetector creates a new ReactDetector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{}
}

// Detect implements the core.Detector interface
func (d *ReactDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	// Check for package.json
	packageJSON, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		if err == fs.ErrNotExist {
			return core.StackInfo{}, false, nil
		}
		return core.StackInfo{}, false, err
	}

	// Parse package.json
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		return core.StackInfo{}, false, err
	}

	// Check for React dependencies
	hasReact := false
	hasReactDom := false
	for dep := range pkg.Dependencies {
		if dep == "react" {
			hasReact = true
		}
		if dep == "react-dom" {
			hasReactDom = true
		}
	}

	// Must have both react and react-dom
	if !hasReact || !hasReactDom {
		return core.StackInfo{}, false, nil
	}

	// Check for React source files
	entries, err := fs.ReadDir(fsys, "src")
	if err != nil {
		return core.StackInfo{}, false, nil
	}

	hasReactFile := false
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".jsx") || strings.HasSuffix(entry.Name(), ".tsx") || strings.HasSuffix(entry.Name(), ".js")) {
			content, err := fs.ReadFile(fsys, "src/"+entry.Name())
			if err != nil {
				continue
			}
			if strings.Contains(string(content), "import React") {
				hasReactFile = true
				break
			}
		}
	}

	if !hasReactFile {
		return core.StackInfo{}, false, nil
	}

	info := core.StackInfo{
		Name:          "react",
		BuildTool:     "npm",
		DetectedFiles: []string{"package.json"},
	}

	if logSink != nil {
		logs.Tag("detect", "detector=%s found=true path=%s", d.Name(), info.DetectedFiles[0])
	}

	return info, true, nil
}

// Name returns the detector name
func (d *ReactDetector) Name() string {
	return "react"
}

// SetLogSink sets the log sink for the detector
func (d *ReactDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}
