package detectors

import (
	"context"
	"errors"
	"io/fs"
	"sync"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/node"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
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

// FanoutDetector implements core.Detector by running all detectors in parallel
type FanoutDetector struct {
	detectors []core.Detector
	logSink   core.LogSink
}

// NewFanoutDetector creates a new FanoutDetector with all available detectors
func NewFanoutDetector() *FanoutDetector {
	return &FanoutDetector{
		detectors: []core.Detector{
			spring.NewSpringBootDetectorV2(),
			react.NewReactDetector(),
			node.NewNodeDetector(),
		},
	}
}

// Detect implements the core.Detector interface
func (d *FanoutDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink

	results := make(chan detectionResult, len(d.detectors))
	var wg sync.WaitGroup

	for _, detector := range d.detectors {
		wg.Add(1)
		go func(detector core.Detector) {
			defer wg.Done()
			info, found, err := detector.Detect(ctx, fsys, logSink)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				results <- detectionResult{
					detector: detector,
					info:     info,
					found:    found,
					err:      err,
				}
				return
			}
			results <- detectionResult{
				detector: detector,
				info:     info,
				found:    found,
				err:      nil,
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

// Name returns the detector name
func (d *FanoutDetector) Name() string {
	return "fanout"
}

// SetLogSink sets the log sink for the detector
func (d *FanoutDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
	for _, detector := range d.detectors {
		detector.SetLogSink(logSink)
	}
}
