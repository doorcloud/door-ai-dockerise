package detectors

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// SpringBootDetector implements core.Detector for Spring Boot projects
type SpringBootDetector struct {
	logSink core.LogSink
}

// NewSpringBootDetector creates a new Spring Boot detector
func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{}
}

// Detect implements the Detector interface
func (d *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	// Check for pom.xml or build.gradle
	pomXML, err := fs.ReadFile(fsys, "pom.xml")
	if err == nil {
		// Found Maven project
		if containsSpringBoot(string(pomXML)) {
			info := core.StackInfo{
				Name:          "springboot",
				BuildTool:     "maven",
				DetectedFiles: []string{"pom.xml"},
			}

			if logSink != nil {
				logSink.Log("detector=springboot found=true buildtool=maven")
			}

			return info, true, nil
		}
	}

	buildGradle, err := fs.ReadFile(fsys, "build.gradle")
	if err == nil {
		// Found Gradle project
		if containsSpringBoot(string(buildGradle)) {
			info := core.StackInfo{
				Name:          "springboot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle"},
			}

			if logSink != nil {
				logSink.Log("detector=springboot found=true buildtool=gradle")
			}

			return info, true, nil
		}
	}

	buildGradleKts, err := fs.ReadFile(fsys, "build.gradle.kts")
	if err == nil {
		// Found Gradle Kotlin project
		if containsSpringBoot(string(buildGradleKts)) {
			info := core.StackInfo{
				Name:          "springboot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle.kts"},
			}

			if logSink != nil {
				logSink.Log("detector=springboot found=true buildtool=gradle")
			}

			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// containsSpringBoot checks if the build file contains Spring Boot dependencies
func containsSpringBoot(content string) bool {
	return strings.Contains(content, "org.springframework.boot") ||
		strings.Contains(content, "spring-boot")
}

// Name returns the detector name
func (d *SpringBootDetector) Name() string {
	return "springboot"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}
