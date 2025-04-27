package springboot

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// SpringBootDetector implements detection rules for Spring Boot projects
type SpringBootDetector struct {
	logSink core.LogSink
}

// NewSpringBootDetector creates a new Spring Boot detector
func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{}
}

// Name returns the detector name
func (d *SpringBootDetector) Name() string {
	return "springboot"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	// Check for pom.xml
	pomXml, err := fs.ReadFile(fsys, "pom.xml")
	if err != nil {
		return core.StackInfo{}, false, nil
	}

	// Check for Spring Boot in pom.xml
	if !strings.Contains(string(pomXml), "spring-boot") {
		return core.StackInfo{}, false, nil
	}

	if d.logSink != nil {
		d.logSink.Log("detector=springboot found=true")
	}

	return core.StackInfo{
		Name:      "springboot",
		BuildTool: "maven",
		DetectedFiles: []string{
			"pom.xml",
		},
	}, true, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
