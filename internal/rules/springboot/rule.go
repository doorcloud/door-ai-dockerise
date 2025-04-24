package springboot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
)

// Rule implements the Rule interface for Spring Boot projects.
type Rule struct {
	logger    *slog.Logger
	llmClient *llm.Client
	config    *rules.RuleConfig
}

// NewRule creates a new Spring Boot rule.
func NewRule(logger *slog.Logger, llmClient *llm.Client, config *rules.RuleConfig) *Rule {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	return &Rule{
		logger:    logger,
		llmClient: llmClient,
		config:    config,
	}
}

// Detect determines if the given path contains a Spring Boot project.
func (r *Rule) Detect(path string) bool {
	files, err := r.findFiles(path, "pom.xml")
	if err != nil {
		return false
	}
	return len(files) > 0
}

// Snippets returns relevant code snippets from the Spring Boot project.
func (r *Rule) Snippets(path string) ([]snippet.T, error) {
	pomFiles, err := r.findFiles(path, "pom.xml")
	if err != nil {
		return nil, fmt.Errorf("finding pom.xml: %w", err)
	}

	var snippets []snippet.T
	for _, pomFile := range pomFiles {
		snip, err := snippet.ReadFile(pomFile)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", pomFile, err)
		}
		snippets = append(snippets, snip)
	}

	return snippets, nil
}

// Facts extracts Spring Boot specific facts from the given snippets.
func (r *Rule) Facts(ctx context.Context, snippets []snippet.T, client facts.LLMClient) (facts.Facts, error) {
	var content []string
	for _, s := range snippets {
		content = append(content, s.Content)
	}

	factsMap, err := client.GenerateFacts(ctx, content)
	if err != nil {
		return facts.Facts{}, fmt.Errorf("extracting facts: %w", err)
	}

	f, err := facts.FromJSON(factsMap)
	if err != nil {
		return facts.Facts{}, fmt.Errorf("converting facts: %w", err)
	}

	return f, nil
}

// Dockerfile generates a Dockerfile for the project.
func (r *Rule) Dockerfile(ctx context.Context, f facts.Facts, llmClient facts.LLMClient) (string, error) {
	factsMap := f.ToMap()
	dockerfile, err := llmClient.GenerateDockerfile(ctx, factsMap)
	if err != nil {
		return "", fmt.Errorf("generating dockerfile: %w", err)
	}

	return dockerfile, nil
}

// Register the Spring Boot rule
func init() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := &rules.RuleConfig{}
	rule := NewRule(logger, nil, config)
	rules.Register(rule)
}

// findFiles finds all files matching the given patterns in the directory.
func (r *Rule) findFiles(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
