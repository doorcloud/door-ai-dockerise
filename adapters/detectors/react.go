package detectors

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
)

type React struct {
	d react.ReactDetector
}

func NewReact() *React {
	return &React{d: react.ReactDetector{}}
}

func (r *React) Detect(ctx context.Context, dir string) (core.StackInfo, error) {
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
