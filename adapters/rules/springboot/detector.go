package springboot

import (
	"io/fs"
	"path/filepath"
)

// SpringBootDetector implements detection rules for Spring Boot projects
type SpringBootDetector struct{}

// Detect checks if the given filesystem contains a Spring Boot project
func (d SpringBootDetector) Detect(fsys fs.FS) bool {
	// Check for pom.xml or build.gradle
	hasBuildFile := false
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() && (filepath.Base(path) == "pom.xml" || filepath.Base(path) == "build.gradle") {
			hasBuildFile = true
			return fs.SkipDir
		}
		return nil
	})

	if !hasBuildFile {
		return false
	}

	// Check for Spring Boot annotations in Java files
	hasSpringBoot := false
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() && filepath.Ext(path) == ".java" {
			content, err := fs.ReadFile(fsys, path)
			if err == nil && (contains(string(content), "@SpringBootApplication") || contains(string(content), "spring-boot-starter")) {
				hasSpringBoot = true
				return fs.SkipDir
			}
		}
		return nil
	})

	return hasSpringBoot
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
