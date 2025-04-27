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

type DetectorOptions struct {
	LogSink io.Writer
}

type ParallelDetector struct {
	detectors []core.Detector
	opts      *DetectorOptions
}

func (p *ParallelDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	results := make(chan detectorResult, len(p.detectors))
	var wg sync.WaitGroup

	for _, d := range p.detectors {
		wg.Add(1)
		go func(detector core.Detector) {
			defer wg.Done()
			info, found, err := detector.Detect(ctx, fsys)
			if p.opts != nil && p.opts.LogSink != nil && found {
				fmt.Fprintf(p.opts.LogSink, "detector=%s found=%v path=%s\n",
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
func NewParallelDetector(detectors []core.Detector, opts *DetectorOptions) *ParallelDetector {
	if opts == nil {
		opts = &DetectorOptions{}
	}
	return &ParallelDetector{
		detectors: detectors,
		opts:      opts,
	}
}

// Name returns the detector name
func (p *ParallelDetector) Name() string {
	return "parallel"
}
