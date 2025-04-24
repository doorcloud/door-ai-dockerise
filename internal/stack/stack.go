package stack

import (
	"context"
	"io/fs"
)

// Rule represents a technology stack detection rule
type Rule interface {
	// Detect returns true if the rule matches the given filesystem
	Detect(ctx context.Context, fsys fs.FS) (bool, error)
	// Snippets returns relevant file fragments for analysis
	Snippets(ctx context.Context, fsys fs.FS) ([]string, error)
	// Language returns the primary language of the stack
	Language() string
	// Framework returns the framework name if any
	Framework() string
	// BaseImageHint returns a suggested base image
	BaseImageHint() string
	// ExposedPorts returns the ports that should be exposed
	ExposedPorts() []int
	// FactsPrompt builds the prompt for fact extraction
	FactsPrompt(snippets []string) string
	// DockerfilePrompt builds the prompt for Dockerfile generation
	DockerfilePrompt(facts Facts, currentDF string, lastErr string) string
}

// Registry is a collection of rules
type Registry []Rule

// Match returns the first rule that detects the given filesystem
func (r Registry) Match(ctx context.Context, fsys fs.FS) (Rule, error) {
	for _, rule := range r {
		matches, err := rule.Detect(ctx, fsys)
		if err != nil {
			return nil, err
		}
		if matches {
			return rule, nil
		}
	}
	return nil, nil
}

// Facts represents the detected facts about a technology stack
type Facts struct {
	Language  string            // "java", "node", "python"…
	Framework string            // "spring-boot", "express", "flask"…
	BuildTool string            // "maven", "npm", "pip", …
	BuildCmd  string            // e.g. "mvn package", "npm run build"
	BuildDir  string            // directory containing build files (e.g. ".", "backend/")
	StartCmd  string            // e.g. "java -jar app.jar", "node server.js"
	Artifact  string            // glob or relative path
	Ports     []int             // e.g. [8080], [3000]
	Health    string            // URL path or CMD
	Env       map[string]string // e.g. {"NODE_ENV": "production"}
	BaseImage string            // e.g. "eclipse-temurin:17-jdk"
	DevMode   bool              // whether to include development dependencies
}

// Detect tries to detect the technology stack in the given filesystem
func Detect(ctx context.Context, fsys fs.FS) (Rule, error) {
	registry := Registry{
		NewSpringBootRule(),
	}
	return registry.Match(ctx, fsys)
}
