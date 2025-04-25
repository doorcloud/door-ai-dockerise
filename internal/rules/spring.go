package rules

import (
	"io/fs"
	"strings"
)

// springDetector implements the types.Detector interface
type springDetector struct{}

// Detect checks if the project is a Spring Boot project
func (d *springDetector) Detect(fsys fs.FS) (bool, error) {
	// Check for pom.xml
	_, err := fs.Stat(fsys, "pom.xml")
	if err != nil {
		return false, nil
	}

	// Check for Spring Boot dependencies in pom.xml
	pomContent, err := fs.ReadFile(fsys, "pom.xml")
	if err != nil {
		return false, err
	}

	return strings.Contains(string(pomContent), "spring-boot-starter"), nil
}

// Name returns the name of the detector
func (d *springDetector) Name() string {
	return "spring"
}
