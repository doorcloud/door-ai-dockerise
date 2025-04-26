package legacyshim

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// DetectorWrapper wraps the new core.Detector interface to match the old detector interface
type DetectorWrapper struct {
	detector core.Detector
}

// NewDetectorWrapper creates a new wrapper for the old detector interface
func NewDetectorWrapper() *DetectorWrapper {
	return &DetectorWrapper{
		detector: react.NewReactDetector(),
	}
}

// Detect implements the old detector interface
func (w *DetectorWrapper) Detect(ctx context.Context, path string) (core.StackInfo, error) {
	return w.detector.Detect(ctx, path)
}
