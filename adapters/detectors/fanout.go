package detectors

import (
	"context"
	"io/fs"
	"sync"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/logs"
)

// LogSink is an interface for logging messages
type LogSink interface {
	Log(msg string)
}

// detectionResult represents the result of a detection attempt
type detectionResult struct {
	detector core.Detector
	info     core.StackInfo
	found    bool
	err      error
}

// ParallelDetector implements Detector by trying all detectors in parallel
type ParallelDetector struct {
	detectors []core.Detector
}

// NewParallelDetector creates a new ParallelDetector
func NewParallelDetector(detectors ...core.Detector) *ParallelDetector {
	return &ParallelDetector{
		detectors: detectors,
	}
}

// Detect implements the Detector interface for ParallelDetector
func (d *ParallelDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	results := make(chan detectionResult, len(d.detectors))
	var wg sync.WaitGroup

	for _, detector := range d.detectors {
		wg.Add(1)
		go func(detector core.Detector) {
			defer wg.Done()
			info, found, err := detector.Detect(ctx, fsys, logSink)
			results <- detectionResult{
				detector: detector,
				info:     info,
				found:    found,
				err:      err,
			}
		}(detector)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		if result.found {
			if logSink != nil {
				logs.Tag("detect", "Detected stack: %s using %s", result.info.Name, result.detector.Name())
			}
			return result.info, true, nil
		}
	}

	if firstErr != nil {
		return core.StackInfo{}, false, firstErr
	}
	return core.StackInfo{}, false, nil
}

// Name returns the name of the detector
func (d *ParallelDetector) Name() string {
	return "parallel"
}
