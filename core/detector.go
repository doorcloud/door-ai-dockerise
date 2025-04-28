package core

import (
	"context"
	"io/fs"
	"sync"
)

// Detector is the interface that all technology stack detectors must implement.
type Detector interface {
	// Detect returns information about the detected technology stack.
	// The second return value indicates whether a stack was detected.
	Detect(ctx context.Context, fsys fs.FS, logSink LogSink) (StackInfo, bool, error)
	// Name returns the name of the detector.
	Name() string
	// Describe returns a human-readable description of what the detector looks for.
	Describe() string
	// SetLogSink sets the log sink for the detector.
	SetLogSink(logSink LogSink)
}

var (
	detectorsMu sync.RWMutex
	detectors   = make(map[string]Detector)
)

// RegisterDetector registers a detector with the core package.
func RegisterDetector(d Detector) {
	detectorsMu.Lock()
	defer detectorsMu.Unlock()
	detectors[d.Name()] = d
}

// Detect runs all registered detectors on the given filesystem.
func Detect(ctx context.Context, fsys fs.FS, logSink LogSink) (StackInfo, bool, error) {
	detectorsMu.RLock()
	defer detectorsMu.RUnlock()

	for _, d := range detectors {
		info, found, err := d.Detect(ctx, fsys, logSink)
		if err != nil {
			return StackInfo{}, false, err
		}
		if found {
			return info, true, nil
		}
	}

	return StackInfo{}, false, nil
}
