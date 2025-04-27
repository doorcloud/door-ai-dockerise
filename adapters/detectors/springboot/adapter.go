package springboot

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/core"
)

type SpringBootDetector struct {
	d springboot.SpringBootDetector
}

func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{d: springboot.SpringBootDetector{}}
}

func (s *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	if s.d.Detect(fsys) {
		return core.StackInfo{
			Name:          "springboot",
			BuildTool:     "maven",
			DetectedFiles: []string{"pom.xml"},
		}, true, nil
	}
	return core.StackInfo{}, false, nil
}

// Name returns the detector name
func (s *SpringBootDetector) Name() string {
	return "springboot"
}
