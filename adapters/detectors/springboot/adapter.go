package springboot

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// SpringBootDetector implements detection rules for Spring Boot projects
type SpringBootDetector struct {
	detector *springboot.SpringBootDetector
}

// NewSpringBootDetector creates a new Spring Boot detector
func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{
		detector: springboot.NewSpringBootDetector(),
	}
}

// Name returns the detector name
func (d *SpringBootDetector) Name() string {
	return d.detector.Name()
}

// Describe returns a description of what the detector looks for
func (d *SpringBootDetector) Describe() string {
	return d.detector.Describe()
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetector) SetLogSink(logSink core.LogSink) {
	d.detector.SetLogSink(logSink)
}

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	return d.detector.Detect(ctx, fsys, logSink)
}
