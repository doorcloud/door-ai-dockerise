package node

import (
	"context"
	"log/slog"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
	"github.com/doorcloud/door-ai-dockerise/pkg/rule"
)

type Rule struct {
	rule.BaseRule
}

func New(logger *slog.Logger) *Rule {
	return &Rule{
		BaseRule: rule.NewBaseRule(logger),
	}
}

func (r *Rule) Detect(path string) bool {
	return r.DetectAny(path, "package.json")
}

func (r *Rule) Snippets(path string) ([]snippet.T, error) {
	// TODO: Implement snippet extraction
	return nil, nil
}

func (r *Rule) Facts(ctx context.Context, snips []snippet.T, c *llm.Client) (facts.Facts, error) {
	// TODO: Implement facts extraction
	return facts.Facts{}, nil
}

func (r *Rule) Dockerfile(ctx context.Context, f facts.Facts, c *llm.Client) (string, error) {
	// TODO: Implement Dockerfile generation
	return "", nil
}
