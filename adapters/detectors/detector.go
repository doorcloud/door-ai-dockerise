package detectors

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Detector defines the interface for project type detection
type Detector interface {
	// Detect attempts to detect the stack type in the given filesystem
	Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error)
	// Name returns the name of the detector
	Name() string
	// SetLogSink sets the log sink for the detector
	SetLogSink(logSink core.LogSink)
}
