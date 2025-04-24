package rules

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
)

// Rule defines the interface for project detection and Dockerfile generation.
type Rule interface {
	// Detect determines if this rule applies to the given path.
	Detect(path string) bool

	// Snippets extracts relevant code snippets from the project.
	Snippets(path string) ([]snippet.T, error)

	// Facts extracts project-specific facts from the given snippets.
	Facts(ctx context.Context, snippets []snippet.T, client facts.LLMClient) (facts.Facts, error)

	// Dockerfile generates a Dockerfile for the project.
	Dockerfile(ctx context.Context, f facts.Facts, c facts.LLMClient) (string, error)
}

// BaseRule provides common functionality for all rules.
type BaseRule struct {
	logger *slog.Logger
}

// NewBaseRule creates a new BaseRule with the given logger.
func NewBaseRule(logger *slog.Logger) BaseRule {
	return BaseRule{logger: logger}
}

// Log returns the rule's logger.
func (r BaseRule) Log() *slog.Logger {
	return r.logger
}

// DetectSentinelFiles checks if any of the given files exist in the path.
func (r BaseRule) DetectSentinelFiles(path string, files []string) bool {
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			return true
		}
	}
	return false
}

// FindFiles returns all files matching the given glob patterns.
func (r BaseRule) FindFiles(path string, patterns []string) ([]string, error) {
	var matches []string
	for _, pattern := range patterns {
		glob := filepath.Join(path, pattern)
		files, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", glob, err)
		}
		matches = append(matches, files...)
	}
	return matches, nil
}

// RegisteredRules holds all registered rules.
var RegisteredRules []Rule

// Register adds a rule to the registry.
func Register(r Rule) {
	RegisteredRules = append(RegisteredRules, r)
}
