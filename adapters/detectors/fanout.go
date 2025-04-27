package detectors

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type detectorResult struct {
	info  core.StackInfo
	found bool
	err   error
}

type ParallelDetector struct {
	detectors []core.Detector
	logSink   io.Writer
}

func (p *ParallelDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	results := make(chan detectorResult, len(p.detectors))
	var wg sync.WaitGroup

	for _, d := range p.detectors {
		wg.Add(1)
		go func(detector core.Detector) {
			defer wg.Done()
			info, found, err := detector.Detect(ctx, fsys)
			if p.logSink != nil && found {
				fmt.Fprintf(p.logSink, "detector=%s found=%v path=%s\n",
					detector.Name(), found, info.DetectedFiles[0])
			}
			results <- detectorResult{info: info, found: found, err: err}
		}(d)
	}

	wg.Wait()
	close(results)

	for result := range results {
		if result.err != nil {
			return core.StackInfo{}, false, result.err
		}
		if result.found {
			return result.info, true, nil
		}
	}
	return core.StackInfo{}, false, nil
}

// NewParallelDetector creates a new ParallelDetector instance
func NewParallelDetector(detectors []core.Detector) *ParallelDetector {
	return &ParallelDetector{
		detectors: detectors,
	}
}

// Name returns the detector name
func (p *ParallelDetector) Name() string {
	return "parallel"
}

// SetLogSink sets the log sink for the detector
func (p *ParallelDetector) SetLogSink(w io.Writer) {
	p.logSink = w
	// Propagate log sink to child detectors
	for _, d := range p.detectors {
		d.SetLogSink(w)
	}
}
