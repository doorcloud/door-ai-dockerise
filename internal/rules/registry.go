package rules

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/doorcloud/door-ai-dockerise/pkg/rule"
)

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

// RegisterRule adds a new rule to the registry
func RegisterRule(name string, r rule.Rule) {
	rule.RegisterDefault(name, r)
}

// DetectRule tries to find a matching rule for the given repository
func DetectRule(repo string) (*StackRule, Facts, error) {
	// Try each registered rule in order
	rule, err := FindStackRule(repo)
	if err != nil {
		return nil, Facts{}, err
	}
	if rule == nil {
		return nil, Facts{}, ErrNoRule
	}

	// Extract facts from the rule
	facts := Facts{
		Language:  rule.BuildHints["language"],
		Framework: rule.Name,
		BuildTool: rule.BuildHints["builder"],
		BuildDir:  rule.BuildHints["build_dir"],
	}

	return rule, facts, nil
}
