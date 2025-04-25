package springboot

import (
	"io/fs"
)

// Detector implements types.Detector.
type Detector struct{}

func (d Detector) Name() string {
	return "spring-boot"
}

func (d Detector) Detect(fsys fs.FS) (bool, error) {
	// Check for Maven
	if _, err := fs.Stat(fsys, "pom.xml"); err == nil {
		return true, nil
	}

	// Check for Gradle
	if _, err := fs.Stat(fsys, "gradlew"); err == nil {
		return true, nil
	}

	// Check for Gradle Kotlin
	if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		return true, nil
	}

	return false, nil
}

// Rule implements types.Rule.
type Rule struct{}

func (r Rule) Name() string {
	return "spring-boot"
}

func (r Rule) Detect(fsys fs.FS) bool {
	// Check for Maven
	if _, err := fs.Stat(fsys, "pom.xml"); err == nil {
		return true
	}

	// Check for Gradle
	if _, err := fs.Stat(fsys, "gradlew"); err == nil {
		return true
	}

	// Check for Gradle Kotlin
	if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		return true
	}

	return false
}

func (r Rule) Facts(fsys fs.FS) map[string]any {
	// Determine build tool
	tool := "maven"
	if _, err := fs.Stat(fsys, "gradlew"); err == nil {
		tool = "gradle"
	} else if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		tool = "gradle"
	}

	return map[string]any{
		"language":   "Java",
		"framework":  "Spring Boot",
		"build_tool": tool,
		"build_cmd":  "mvn clean package",
		"start_cmd":  "java -jar target/*.jar",
		"artifact":   "target/*.jar",
		"ports":      []int{8080},
	}
}
