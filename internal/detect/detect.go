package detect

import (
	"io/fs"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// RuleInfo represents information about a detected technology stack
type RuleInfo struct {
	Name string // e.g. "spring-boot"
	Tool string // e.g. "maven"
}

// Detect checks if the given filesystem matches any known rules
func Detect(path fs.FS) (RuleInfo, error) {
	// Check for Spring Boot with Maven
	exists, err := fs.Stat(path, "pom.xml")
	if err == nil && !exists.IsDir() {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "maven",
		}, nil
	}

	// Check for Gradle wrapper
	exists, err = fs.Stat(path, "gradlew")
	if err == nil && !exists.IsDir() {
		return RuleInfo{
			Name: "spring-boot",
			Tool: "gradle",
		}, nil
	}

	// Check for pnpm
	exists, err = fs.Stat(path, "pnpm-lock.yaml")
	if err == nil && !exists.IsDir() {
		return RuleInfo{
			Name: "node",
			Tool: "pnpm",
		}, nil
	}

	return RuleInfo{}, nil
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
func DetectProjectType(fsys fs.FS, registry types.Registry) (string, error) {
	detectors := registry.GetDetectors()
	for _, detector := range detectors {
		detected, err := detector.Detect(fsys)
		if err != nil {
			return "", err
		}
		if detected {
			return detector.Name(), nil
		}
	}
	return "", nil
}
