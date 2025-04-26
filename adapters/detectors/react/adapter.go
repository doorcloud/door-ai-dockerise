package react

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/react"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// ReactDetector implements core.Detector for React projects
type ReactDetector struct {
	d react.ReactDetector
}

// NewReactDetector creates a new ReactDetector
func NewReactDetector() *ReactDetector {
	return &ReactDetector{d: react.ReactDetector{}}
}

// Detect implements the core.Detector interface
func (r *ReactDetector) Detect(ctx context.Context, dir string) (core.StackInfo, error) {
	fsys := os.DirFS(dir)
	if r.d.Detect(fsys) {
		return core.StackInfo{
			Name: "react",
			Meta: map[string]string{
				"framework": "react",
			},
		}, nil
	}
	return core.StackInfo{}, nil
}
