package detectors

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/react"
	"github.com/doorcloud/door-ai-dockerise/core"
)

type React struct {
	d react.ReactDetector
}

func NewReact() *React {
	return &React{d: react.ReactDetector{}}
}

func (r *React) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, error) {
	if r.d.Detect(fsys) {
		return core.StackInfo{
			Name:      "react",
			BuildTool: "npm",
		}, nil
	}
	return core.StackInfo{}, nil
}
