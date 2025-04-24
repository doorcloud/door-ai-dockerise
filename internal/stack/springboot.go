package stack

import (
	"context"
	"fmt"
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

// BaseImageHint returns a suggested base image
func (r *SpringBootRule) BaseImageHint() string {
	return "eclipse-temurin:17-jdk"
}

// ExposedPorts returns the ports that should be exposed
func (r *SpringBootRule) ExposedPorts() []int {
	return []int{8080}
}

// FactsPrompt builds the prompt for fact extraction
func (r *SpringBootRule) FactsPrompt(snippets []string) string {
	return fmt.Sprintf(`You are a code analysis expert. Given a set of code snippets, extract key facts about the project.
The output must be valid JSON with the following structure:
{
  "language": "java",
  "framework": "spring-boot",
  "build_tool": "build system (maven, gradle, etc)",
  "build_cmd": "command to build",
  "build_dir": "directory containing build files (e.g., '.', 'backend/')",
  "start_cmd": "command to start the application",
  "artifact": "path to built artifact",
  "ports": [8080],
  "env": {"key": "value"},
  "health": "/actuator/health"
}

Code snippets:
%s`, snippets)
}

// DockerfilePrompt builds the prompt for Dockerfile generation
func (r *SpringBootRule) DockerfilePrompt(facts Facts, currentDF string, lastErr string) string {
	prompt := fmt.Sprintf(`You are a Docker expert. Create a production-ready Dockerfile for a %s application using %s.
Facts about the application:
- Language: %s
- Framework: %s
- Build tool: %s
- Build command: %s
- Start command: %s
- Ports: %v
- Health check: %s
- Base image: %s

Requirements:
- Use multi-stage build
- Optimize for production
- Include health check
- Set appropriate labels
- Use non-root user
- Handle environment variables
- Include proper error handling

The Dockerfile should be valid and buildable.`, facts.Language, facts.Framework, facts.Language, facts.Framework,
		facts.BuildTool, facts.BuildCmd, facts.StartCmd, facts.Ports, facts.Health, facts.BaseImage)

	if currentDF != "" {
		prompt += fmt.Sprintf(`

Previous Dockerfile that failed:
%s

Error:
%s

Please fix the issues while maintaining the working parts.`, currentDF, lastErr)
	}

	return prompt
}
