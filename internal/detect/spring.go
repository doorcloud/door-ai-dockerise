package detect

import (
	"io/fs"
)

// SpringDetector implements the Detector interface for Spring Boot projects
type SpringDetector struct{}

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringDetector) Detect(fsys fs.FS) (RuleInfo, bool) {
	// Check for pom.xml
	exists, err := fs.Stat(fsys, "pom.xml")
	if err == nil && !exists.IsDir() {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "maven",
		}, true
	}

	// Check for Gradle wrapper
	exists, err = fs.Stat(fsys, "gradlew")
	if err == nil && !exists.IsDir() {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "gradle",
		}, true
	}

	return RuleInfo{}, false
}
