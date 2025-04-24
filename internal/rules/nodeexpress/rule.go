package nodeexpress

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
	"github.com/doorcloud/door-ai-dockerise/pkg/rule"
)

type NodeExpressRule struct {
	rule.BaseRule
}

func init() {
	rule.RegisterDefault("nodeexpress", &NodeExpressRule{
		BaseRule: rule.NewBaseRule(slog.Default()),
	})
}

func (r *NodeExpressRule) Detect(path string) bool {
	// Check for package.json
	matches, err := doublestar.Glob(os.DirFS(path), "**/package.json")
	if err != nil {
		return false
	}
	if len(matches) == 0 {
		return false
	}

	// Parse package.json to check for express dependency
	pkgFile := filepath.Join(path, matches[0])
	content, err := os.ReadFile(pkgFile)
	if err != nil {
		return false
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return false
	}

	// Check for express in dependencies
	for dep := range pkg.Dependencies {
		if dep == "express" {
			return true
		}
	}
	return false
}

func (r *NodeExpressRule) Snippets(path string) ([]snippet.T, error) {
	// TODO: Implement snippet extraction
	return nil, nil
}

func (r *NodeExpressRule) Facts(ctx context.Context, snips []snippet.T, c *llm.Client) (facts.Facts, error) {
	return facts.Facts{
		Language:  "node",
		Framework: "express",
		BuildTool: "npm",
		BuildCmd:  "npm install",
		Artifact:  ".",
		Ports:     []int{3000},
		Health:    "/health",
		BaseHint:  "node:20-alpine",
	}, nil
}

func (r *NodeExpressRule) Dockerfile(ctx context.Context, f facts.Facts, c *llm.Client) (string, error) {
	return `FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build --if-present

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/dist ./dist
EXPOSE 3000
CMD ["node", "dist/index.js"]`, nil
}
