package react

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
)

type ReactDetector struct {
	detector react.ReactDetector
}

func NewReactDetector() *ReactDetector {
	return &ReactDetector{
		detector: react.ReactDetector{},
	}
}

func (d *ReactDetector) Detect(ctx context.Context, path string) (core.StackInfo, error) {
	fsys := os.DirFS(path)
	if d.detector.Detect(fsys) {
		facts := d.detector.Facts()
		return core.StackInfo{
			Name: "react",
			Meta: map[string]string{
				"framework": "react",
				"buildTool": facts.BuildTool,
			},
		}, nil
	}
	return core.StackInfo{}, nil
}
