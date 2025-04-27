package react

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// ReactDetector implements the core.Detector interface for React projects
type ReactDetector struct {
	logSink io.Writer
}

// NewReactDetector creates a new ReactDetector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{}
}

// Detect implements the core.Detector interface
func (d *ReactDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	// Check for package.json
	packageJSON, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		if err == fs.ErrNotExist {
			return core.StackInfo{}, false, nil
		}
		return core.StackInfo{}, false, err
	}

	// Check for React dependencies
	if !containsReact(string(packageJSON)) {
		return core.StackInfo{}, false, nil
	}

	info := core.StackInfo{
		Name:          "react",
		BuildTool:     "npm",
		DetectedFiles: []string{"package.json"},
	}

	if d.logSink != nil {
		io.WriteString(d.logSink, fmt.Sprintf("detector=%s found=true path=%s\n", d.Name(), info.DetectedFiles[0]))
	}

	return info, true, nil
}

// containsReact checks if the package.json contains React dependencies
func containsReact(packageJSON string) bool {
	return strings.Contains(packageJSON, `"react"`) || strings.Contains(packageJSON, `"@types/react"`)
}

// Name returns the detector name
func (d *ReactDetector) Name() string {
	return "react"
}

// SetLogSink sets the log sink for the detector
func (d *ReactDetector) SetLogSink(w io.Writer) {
	d.logSink = w
}
