package springboot

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/core"
)

type SpringBootDetector struct {
	d springboot.SpringBootDetector
}

func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{d: springboot.SpringBootDetector{}}
}

func (s *SpringBootDetector) Detect(ctx context.Context, dir string) (core.StackInfo, error) {
	fsys := os.DirFS(dir)
	if s.d.Detect(fsys) {
		return core.StackInfo{
			Name: "springboot",
			Meta: map[string]string{
				"framework": "springboot",
				"language":  "java",
			},
		}, nil
	}
	return core.StackInfo{}, nil
}
