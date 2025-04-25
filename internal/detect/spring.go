package detect

import (
	"io/fs"
)

// SpringDetector implements detection for Spring Boot projects
type SpringDetector struct{}

// Detect checks for Spring Boot project markers
func (d *SpringDetector) Detect(fsys fs.FS) (Rule, bool) {
	// Check for Maven
	exists, err := fs.Stat(fsys, "pom.xml")
	if err == nil && !exists.IsDir() {
		return Rule{
			Name: "spring-boot",
			Tool: "maven",
		}, true
	}

	// Check for Gradle
	exists, err = fs.Stat(fsys, "gradlew")
	if err == nil && !exists.IsDir() {
		return Rule{
			Name: "spring-boot",
			Tool: "gradle",
		}, true
	}

	return Rule{}, false
}
