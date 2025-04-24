package rules

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// StackRule represents a technology stack rule.
type StackRule struct {
	Name          string            // e.g. "springboot", "nodejs"
	Signatures    []string          // glob patterns that prove the stack
	ManifestGlobs []string          // files we must feed to the LLM
	CodeGlobs     []string          // optional – biggest 2 go to LLM if needed
	MainRegex     string            // marker that identifies the main class/file
	BuildHints    map[string]string // e.g. {"builder":"maven:3.9-eclipse-temurin-21"}
}

// DetectFunc is the signature for stack detection functions.
type DetectFunc func(repo string) (*StackRule, error)

var (
	// registeredDetectors holds all registered detection functions
	registeredDetectors []DetectFunc
)

// RegisterDetector adds a new detection function to the registry.
func RegisterDetector(detect DetectFunc) {
	registeredDetectors = append(registeredDetectors, detect)
}

// FindStackRule tries to detect the stack type for a repository.
func FindStackRule(repo string) (*StackRule, error) {
	// Try each registered rule in order
	for _, detect := range registeredDetectors {
		rule, err := detect(repo)
		if err != nil {
			slog.Error("rule detection failed", "error", err)
			continue
		}
		if rule != nil {
			return rule, nil
		}
	}

	// Try the fallback rule
	return fallbackRule(repo), nil
}

// fallbackRule is a last-resort rule that looks for common manifest files.
func fallbackRule(repo string) *StackRule {
	// Try to determine the build directory by looking for common files
	buildDir := "."
	if matches, err := doublestar.Glob(os.DirFS(repo), "**/package.json"); err == nil && len(matches) > 0 {
		buildDir = filepath.Dir(matches[0])
	} else if matches, err := doublestar.Glob(os.DirFS(repo), "**/pom.xml"); err == nil && len(matches) > 0 {
		buildDir = filepath.Dir(matches[0])
	} else if matches, err := doublestar.Glob(os.DirFS(repo), "**/go.mod"); err == nil && len(matches) > 0 {
		buildDir = filepath.Dir(matches[0])
	}

	return &StackRule{
		Name: "fallback",
		Signatures: []string{
			"**/*.js",
			"**/*.py",
			"**/*.rb",
			"**/*.java",
			"**/*.go",
		},
		ManifestGlobs: []string{
			"**/package.json",
			"**/requirements.txt",
			"**/Gemfile",
			"**/pom.xml",
			"**/go.mod",
		},
		CodeGlobs: []string{
			"**/*.js",
			"**/*.py",
			"**/*.rb",
			"**/*.java",
			"**/*.go",
		},
		MainRegex: "main|app|index",
		BuildHints: map[string]string{
			"builder":   "ubuntu:latest",
			"build_dir": buildDir,
		},
	}
}
