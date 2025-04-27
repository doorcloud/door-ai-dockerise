package springboot

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/logs"
)

type SpringBootDetector struct {
	d       springboot.SpringBootDetector
	logSink core.LogSink
}

func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{d: springboot.SpringBootDetector{}}
}

func (s *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	if s.d.Detect(fsys) {
		info := core.StackInfo{
			Name:          "springboot",
			BuildTool:     "maven",
			DetectedFiles: []string{"pom.xml"},
		}

		if logSink != nil {
			logs.Tag("detect", "detector=%s found=true path=%s", s.Name(), info.DetectedFiles[0])
		}

		return info, true, nil
	}
	return core.StackInfo{}, false, nil
}

// Name returns the detector name
func (s *SpringBootDetector) Name() string {
	return "springboot"
}

// SetLogSink sets the log sink for the detector
func (s *SpringBootDetector) SetLogSink(logSink core.LogSink) {
	s.logSink = logSink
}
