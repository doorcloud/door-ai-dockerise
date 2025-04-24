package python

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
)

// Rule implements the Rule interface for Python projects.
type Rule struct {
	rules.BaseRule
}

// NewRule creates a new Python rule.
func NewRule(logger *slog.Logger) *Rule {
	return &Rule{
		BaseRule: rules.NewBaseRule(logger),
	}
}

// Detect returns true if this is a Python project.
func (r *Rule) Detect(path string) bool {
	// Fast: Check for Python files and requirements
	if !r.DetectSentinelFiles(path, []string{"requirements.txt", "pyproject.toml", "setup.py"}) {
		return false
	}

	// Medium: Check for framework-specific files
	frameworkFiles := []string{
		"manage.py",   // Django
		"app.py",      // Flask
		"main.py",     // FastAPI
		"wsgi.py",     // WSGI
		"asgi.py",     // ASGI
		"Pipfile",     // Pipenv
		"poetry.lock", // Poetry
	}
	return r.DetectSentinelFiles(path, frameworkFiles)
}

// Snippets extracts relevant code snippets from the repository.
func (r *Rule) Snippets(path string) ([]snippet.T, error) {
	// Look for key files
	patterns := []string{
		"requirements.txt",
		"pyproject.toml",
		"setup.py",
		"manage.py",
		"app.py",
		"main.py",
		"wsgi.py",
		"asgi.py",
		"Pipfile",
		"poetry.lock",
		"*.py",
	}

	paths, err := r.FindFiles(path, patterns)
	if err != nil {
		return nil, fmt.Errorf("find files: %w", err)
	}

	// Limit Python files to first 120 lines
	snippets, err := snippet.ReadFilesWithLimit(paths, 120)
	if err != nil {
		return nil, fmt.Errorf("read files: %w", err)
	}

	snippet.Log(r.Log(), snippets)
	return snippets, nil
}

// Facts uses the LLM to analyze snippets and extract project facts.
func (r *Rule) Facts(ctx context.Context, snips []snippet.T, client facts.LLMClient) (facts.Facts, error) {
	// Convert snippets to strings for the LLM
	var snippetStrs []string
	for _, s := range snips {
		snippetStrs = append(snippetStrs, fmt.Sprintf("=== %s ===\n%s", s.Path, s.Content))
	}

	// Get facts from LLM
	factsMap, err := client.GenerateFacts(ctx, snippetStrs)
	if err != nil {
		return facts.Facts{}, fmt.Errorf("generate facts: %w", err)
	}

	// Convert to Facts struct
	f, err := facts.FromJSON(factsMap)
	if err != nil {
		return facts.Facts{}, fmt.Errorf("parse facts: %w", err)
	}

	f.Log(r.Log())
	return f, nil
}

// Dockerfile generates a Dockerfile based on the extracted facts.
func (r *Rule) Dockerfile(ctx context.Context, f facts.Facts, client facts.LLMClient) (string, error) {
	// Convert facts to map for LLM
	factsMap := f.ToMap()

	// Generate Dockerfile
	dockerfile, err := client.GenerateDockerfile(ctx, factsMap)
	if err != nil {
		return "", fmt.Errorf("generate dockerfile: %w", err)
	}

	if os.Getenv("DG_DEBUG") == "1" {
		r.Log().Debug("generated dockerfile", "content", dockerfile)
	}

	return dockerfile, nil
}

func init() {
	rules.Register(NewRule(slog.Default()))
}
