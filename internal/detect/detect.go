package detect

import (
	"errors"
	"io/fs"
	"os"
)

// ErrUnknownStack is returned when no rule matches the project
var ErrUnknownStack = errors.New("unknown technology stack")

// RuleInfo represents information about a detected technology stack
type RuleInfo struct {
	Name string // e.g. "spring-boot"
	Tool string // e.g. "maven"
}

// Detect checks if the given filesystem matches any known rules
func Detect(fsys fs.FS) (RuleInfo, error) {
	// Check for Spring Boot with Maven
	if _, err := fs.Stat(fsys, "pom.xml"); err == nil {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "maven",
		}, nil
	}

	// Check for Spring Boot with Gradle
	if _, err := fs.Stat(fsys, "build.gradle"); err == nil {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "gradle",
		}, nil
	}

	// Check for Spring Boot with Gradle Kotlin
	if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "gradle",
		}, nil
	}

	return RuleInfo{}, ErrUnknownStack
}

// DetectStack analyzes the given directory and returns the detected technology stack
func DetectStack(dir string) (string, error) {
	fsys := os.DirFS(dir)
	rule, err := Detect(fsys)
	if err != nil {
		return "", err
	}
	return rule.Name, nil
}

// DetectProjectType detects the project type using the provided registry
func DetectProjectType(fsys fs.FS, registry interface{}) (string, error) {
	rule, err := Detect(fsys)
	if err != nil {
		return "", err
	}
	return rule.Name, nil
}
