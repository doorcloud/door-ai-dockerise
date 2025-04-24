package rule

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
)

// Rule defines the interface for technology stack rules.
type Rule interface {
	// Detect returns true if the given path is a project of this type.
	Detect(path string) bool

	// Snippets extracts relevant code snippets from the repository.
	Snippets(path string) ([]snippet.T, error)

	// Facts analyzes the snippets to extract facts about the project.
	Facts(ctx context.Context, snips []snippet.T, c *llm.Client) (facts.Facts, error)

	// Dockerfile generates a Dockerfile for the project.
	Dockerfile(ctx context.Context, f facts.Facts, c *llm.Client) (string, error)

	// Log returns the logger for this rule.
	Log() *slog.Logger
}

// BaseRule provides common functionality for rules.
type BaseRule struct {
	logger *slog.Logger
}

// NewBaseRule creates a new base rule with the given logger.
func NewBaseRule(logger *slog.Logger) BaseRule {
	return BaseRule{logger: logger}
}

// Log returns the logger for this rule.
func (r BaseRule) Log() *slog.Logger {
	return r.logger
}

// DetectAny returns true if any of the given glob patterns match files in the path.
func (r BaseRule) DetectAny(path string, globs ...string) bool {
	for _, glob := range globs {
		matches, err := doublestar.Glob(os.DirFS(path), glob)
		if err != nil {
			r.logger.Error("error matching glob", "glob", glob, "error", err)
			continue
		}
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

// FindFiles returns a list of files matching the given glob patterns, up to maxDepth.
func (r BaseRule) FindFiles(path string, maxDepth int, skipDirs []string, globs ...string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip specified directories
		for _, skip := range skipDirs {
			if strings.Contains(p, skip) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check depth
		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		depth := strings.Count(rel, string(filepath.Separator))
		if depth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check globs
		for _, glob := range globs {
			matched, err := doublestar.Match(glob, rel)
			if err != nil {
				return err
			}
			if matched && !d.IsDir() {
				matches = append(matches, p)
				break
			}
		}

		return nil
	})

	return matches, err
}
