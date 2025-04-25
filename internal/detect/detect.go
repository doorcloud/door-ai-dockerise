package detect

import (
	"io/fs"
)

// Rule represents a technology stack detection rule
type Rule struct {
	Name string // e.g. "spring-boot"
	Tool string // e.g. "maven"
}

// Detect checks if the given filesystem matches any known rules
func Detect(path fs.FS) (Rule, error) {
	// Check for Spring Boot with Maven
	exists, err := fs.Stat(path, "pom.xml")
	if err == nil && !exists.IsDir() {
		return Rule{
			Name: "spring-boot",
			Tool: "maven",
		}, nil
	}

	// Check for Gradle wrapper
	exists, err = fs.Stat(path, "gradlew")
	if err == nil && !exists.IsDir() {
		return Rule{
			Name: "spring-boot",
			Tool: "gradle",
		}, nil
	}

	// Check for pnpm
	exists, err = fs.Stat(path, "pnpm-lock.yaml")
	if err == nil && !exists.IsDir() {
		return Rule{
			Name: "node",
			Tool: "pnpm",
		}, nil
	}

	return Rule{}, nil
}
