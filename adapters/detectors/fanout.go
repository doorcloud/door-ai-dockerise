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
	info core.StackInfo
	err  error
}

type DetectorOptions struct {
	LogSink io.Writer
}

type ParallelDetector struct {
	detectors []core.Detector
	opts      *DetectorOptions
}

func (p *ParallelDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, error) {
	results := make(chan detectorResult, len(p.detectors))
	var wg sync.WaitGroup

	for _, d := range p.detectors {
		wg.Add(1)
		go func(detector core.Detector) {
			defer wg.Done()
			info, err := detector.Detect(ctx, fsys)
			if p.opts.LogSink != nil {
				fmt.Fprintf(p.opts.LogSink, "detector=%T found=%v\n",
					detector, info.Name != "")
			}
			results <- detectorResult{info: info, err: err}
		}(d)
	}

	wg.Wait()
	close(results)

	for result := range results {
		if result.err != nil {
			return core.StackInfo{}, result.err
		}
		if result.info.Name != "" {
			return result.info, nil
		}
	}
	return core.StackInfo{}, nil
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
