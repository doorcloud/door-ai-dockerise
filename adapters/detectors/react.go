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
