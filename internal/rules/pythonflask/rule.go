package pythonflask

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
	"github.com/doorcloud/door-ai-dockerise/pkg/rule"
)

type PythonFlaskRule struct {
	rule.BaseRule
}

func init() {
	rule.RegisterDefault("pythonflask", &PythonFlaskRule{
		BaseRule: rule.NewBaseRule(slog.Default()),
	})
}

func (r *PythonFlaskRule) Detect(path string) bool {
	// Check for requirements.txt or pyproject.toml
	matches, err := doublestar.Glob(os.DirFS(path), "**/{requirements.txt,pyproject.toml}")
	if err != nil {
		return false
	}
	if len(matches) == 0 {
		return false
	}

	// Check for Flask in requirements.txt
	reqFile := filepath.Join(path, matches[0])
	content, err := os.ReadFile(reqFile)
	if err != nil {
		return false
	}

	// Simple check for Flask dependency
	return strings.Contains(strings.ToLower(string(content)), "flask")
}

func (r *PythonFlaskRule) Snippets(path string) ([]snippet.T, error) {
	// TODO: Implement snippet extraction
	return nil, nil
}

func (r *PythonFlaskRule) Facts(ctx context.Context, snips []snippet.T, c *llm.Client) (facts.Facts, error) {
	return facts.Facts{
		Language:  "python",
		Framework: "flask",
		BuildTool: "pip",
		BuildCmd:  "pip install -r requirements.txt",
		Artifact:  ".",
		Ports:     []int{5000},
		Health:    "/health",
		BaseHint:  "python:3.12-slim",
	}, nil
}

func (r *PythonFlaskRule) Dockerfile(ctx context.Context, f facts.Facts, c *llm.Client) (string, error) {
	return `FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 5000
CMD ["python", "app.py"]`, nil
}
