package rules

import (
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

// springDetector implements Detector for Spring Boot projects
type springDetector struct{}

// Detect checks for Spring Boot project markers
func (d *springDetector) Detect(fsys fs.FS) (detect.Rule, bool) {
	// Check for Maven
	exists, err := fs.Stat(fsys, "pom.xml")
	if err == nil && !exists.IsDir() {
		return detect.Rule{
			Name: "spring-boot",
			Tool: "maven",
		}, true
	}

	// Check for Gradle
	exists, err = fs.Stat(fsys, "gradlew")
	if err == nil && !exists.IsDir() {
		return detect.Rule{
			Name: "spring-boot",
			Tool: "gradle",
		}, true
	}

	return detect.Rule{}, false
}
