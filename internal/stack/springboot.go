package stack

import (
	"context"
	"io/fs"
	"path/filepath"
	"regexp"
)

// SpringBootRule implements the Rule interface for Spring Boot projects
type SpringBootRule struct{}

// NewSpringBootRule creates a new Spring Boot rule
func NewSpringBootRule() Rule {
	return &SpringBootRule{}
}

// Detect returns true if the project appears to be a Spring Boot application
func (r *SpringBootRule) Detect(ctx context.Context, fsys fs.FS) (bool, error) {
	// Check for pom.xml or build.gradle
	hasBuildFile := false
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (d.Name() == "pom.xml" || d.Name() == "build.gradle") {
			hasBuildFile = true
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	if !hasBuildFile {
		return false, nil
	}

	// Check for @SpringBootApplication annotation
	hasSpringBootApp := false
	springBootRegex := regexp.MustCompile(`@SpringBootApplication`)
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (filepath.Ext(path) == ".java" || filepath.Ext(path) == ".kt") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			if springBootRegex.Match(content) {
				hasSpringBootApp = true
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	return hasSpringBootApp, nil
}

// Snippets returns relevant file fragments for analysis
func (r *SpringBootRule) Snippets(ctx context.Context, fsys fs.FS) ([]string, error) {
	var snippets []string

	// Collect relevant files
	files := []string{
		"pom.xml",
		"build.gradle",
		"application.properties",
		"application.yml",
		"application.yaml",
	}

	for _, file := range files {
		content, err := fs.ReadFile(fsys, file)
		if err == nil {
			snippets = append(snippets, string(content))
		}
	}

	// Find and add main application class
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (filepath.Ext(path) == ".java" || filepath.Ext(path) == ".kt") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			if regexp.MustCompile(`@SpringBootApplication`).Match(content) {
				snippets = append(snippets, string(content))
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return snippets, nil
}

// Language returns the primary language
func (r *SpringBootRule) Language() string {
	return "java"
}

// Framework returns the framework name
func (r *SpringBootRule) Framework() string {
	return "spring-boot"
}
