package detectors

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/react"
	"github.com/doorcloud/door-ai-dockerise/core"
)

type React struct {
	d       react.ReactDetector
	logSink io.Writer
}

func NewReact() *React {
	return &React{d: react.ReactDetector{}}
}

func (r *React) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	if r.d.Detect(fsys) {
		info := core.StackInfo{
			Name:          "react",
			BuildTool:     "npm",
			DetectedFiles: []string{"package.json"},
		}

		if r.logSink != nil {
			io.WriteString(r.logSink, fmt.Sprintf("detector=%s found=true path=%s\n", r.Name(), info.DetectedFiles[0]))
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
func (r *React) SetLogSink(w io.Writer) {
	r.logSink = w
}
