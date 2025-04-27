package springboot

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/core"
)

type SpringBootDetector struct {
	d       springboot.SpringBootDetector
	logSink io.Writer
}

func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{d: springboot.SpringBootDetector{}}
}

func (s *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	if s.d.Detect(fsys) {
		info := core.StackInfo{
			Name:          "springboot",
			BuildTool:     "maven",
			DetectedFiles: []string{"pom.xml"},
		}

		if s.logSink != nil {
			io.WriteString(s.logSink, fmt.Sprintf("detector=%s found=true path=%s\n", s.Name(), info.DetectedFiles[0]))
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
func (s *SpringBootDetector) SetLogSink(w io.Writer) {
	s.logSink = w
}
