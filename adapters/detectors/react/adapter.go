package react

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
)

// ReactDetector implements core.Detector for React projects
type ReactDetector struct {
	detector react.ReactDetector
}

// NewReactDetector creates a new ReactDetector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{
		detector: react.ReactDetector{},
	}
}

// Detect implements the core.Detector interface
func (d *ReactDetector) Detect(ctx context.Context, path string) (core.StackInfo, error) {
	fsys := os.DirFS(path)
	if d.detector.Detect(fsys) {
		facts := d.detector.Facts(fsys)
		return core.StackInfo{
			Name: "react",
			Meta: map[string]string{
				"framework": "react",
				"buildTool": facts["buildTool"].(string),
			},
		}, nil
	}
	return core.StackInfo{}, nil
}
