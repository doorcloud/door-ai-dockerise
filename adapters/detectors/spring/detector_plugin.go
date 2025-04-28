package spring

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type springDetector struct {
	logSink core.LogSink
}

func (d *springDetector) Name() string {
	return "spring-boot"
}

func (d *springDetector) Describe() string {
	return "Detects Spring Boot applications by looking for Maven or Gradle build files"
}

func (d *springDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}

func (d *springDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink

	// Check for Maven project
	if d.detectMaven(fsys) {
		return core.StackInfo{
			Name:      "spring-boot",
			BuildTool: "maven",
			DetectedFiles: []string{
				"pom.xml",
			},
		}, true, nil
	}

	// Check for Gradle project
	if d.detectGradle(fsys) {
		return core.StackInfo{
			Name:      "spring-boot",
			BuildTool: "gradle",
			DetectedFiles: []string{
				"build.gradle",
				"build.gradle.kts",
			},
		}, true, nil
	}

	return core.StackInfo{}, false, nil
}

func (d *springDetector) detectMaven(fsys fs.FS) bool {
	// Check for pom.xml
	pomFile, err := fs.ReadFile(fsys, "pom.xml")
	if err != nil {
		return false
	}

	// Check if it's a Spring Boot project
	return strings.Contains(string(pomFile), "spring-boot-starter-parent")
}

func (d *springDetector) detectGradle(fsys fs.FS) bool {
	// Check for build.gradle or build.gradle.kts
	gradleFiles := []string{"build.gradle", "build.gradle.kts"}
	for _, file := range gradleFiles {
		gradleFile, err := fs.ReadFile(fsys, file)
		if err != nil {
			continue
		}

		// Check if it's a Spring Boot project
		if strings.Contains(string(gradleFile), "org.springframework.boot") {
			return true
		}
	}

	return false
}

func init() {
	detector := &springDetector{}
	core.RegisterDetector(detector)
}
