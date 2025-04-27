package detectors

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/react"
	"github.com/doorcloud/door-ai-dockerise/core"
)

type React struct {
	d       *react.ReactDetector
	logSink core.LogSink
}

func NewReact() *React {
	return &React{d: react.NewReactDetector()}
}

func (r *React) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	info, found, err := r.d.Detect(ctx, fsys, logSink)
	if err != nil {
		return core.StackInfo{}, false, err
	}
	if found {
		if r.logSink != nil {
			r.logSink.Log(fmt.Sprintf("detector=%s found=true path=%s", r.Name(), info.DetectedFiles[0]))
		}
		return info, true, nil
	}
	return core.StackInfo{}, false, nil
}

// Name returns the detector name
func (r *React) Name() string {
	return "react"
}

// SetLogSink sets the log sink for the detector
func (r *React) SetLogSink(logSink core.LogSink) {
	r.logSink = logSink
	r.d.SetLogSink(logSink)
}

// writerLogSink adapts an io.Writer to a core.LogSink
type writerLogSink struct {
	w io.Writer
}

func (w *writerLogSink) Log(msg string) {
	if w.w != nil {
		io.WriteString(w.w, msg)
	}
}

// ReactDetector implements core.Detector for React projects
type ReactDetector struct {
	logSink core.LogSink
}

// NewReactDetector creates a new React detector
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

	// Check for React dependencies
	content := string(packageJSON)
	if !containsReact(content) {
		return core.StackInfo{}, false, nil
	}

	info := core.StackInfo{
		Name:          "react",
		BuildTool:     "npm",
		DetectedFiles: []string{"package.json"},
	}

	if logSink != nil {
		logSink.Log("detector=react found=true")
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
func (d *ReactDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}
