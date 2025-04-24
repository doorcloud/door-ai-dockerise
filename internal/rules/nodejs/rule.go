package nodejs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/snippet"
	"github.com/doorcloud/door-ai-dockerise/internal/util"
)

// Rule implements the Rule interface for Node.js projects.
type Rule struct {
	rules.BaseRule
	llmClient *llm.Client
}

// NewRule creates a new Node.js rule.
func NewRule(logger *slog.Logger, llmClient *llm.Client) *Rule {
	return &Rule{
		BaseRule:  rules.NewBaseRule(logger),
		llmClient: llmClient,
	}
}

// Detect returns true if this is a Node.js project.
func (r *Rule) Detect(path string) bool {
	// Fast: Check for package.json
	if !r.DetectSentinelFiles(path, []string{"package.json"}) {
		return false
	}

	// Medium: Check for framework-specific files
	frameworkFiles := []string{
		"server.js",
		"app.js",
		"index.js",
		"vite.config.*",
		"angular.json",
		"vue.config.js",
		"next.config.js",
		"nuxt.config.js",
	}
	return r.DetectSentinelFiles(path, frameworkFiles)
}

// Snippets extracts relevant code snippets from the repository.
func (r *Rule) Snippets(path string) ([]snippet.T, error) {
	var snippets []snippet.T

	// Extract package.json
	if util.FileExists(filepath.Join(path, "package.json")) {
		content, err := os.ReadFile(filepath.Join(path, "package.json"))
		if err != nil {
			return nil, fmt.Errorf("read package.json: %w", err)
		}
		snippets = append(snippets, snippet.T{
			Path:    "package.json",
			Content: string(content),
		})
	}

	// Extract main application file
	mainFiles := []string{"server.js", "app.js", "index.js"}
	for _, file := range mainFiles {
		if util.FileExists(filepath.Join(path, file)) {
			content, err := os.ReadFile(filepath.Join(path, file))
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", file, err)
			}
			snippets = append(snippets, snippet.T{
				Path:    file,
				Content: string(content),
			})
			break
		}
	}

	return snippets, nil
}

// Facts analyzes the snippets to extract facts about the project.
func (r *Rule) Facts(ctx context.Context, snips []snippet.T, c *llm.Client) (facts.Facts, error) {
	f := facts.Facts{
		Language:     "javascript",
		Framework:    "nodejs",
		BuildTool:    "npm",
		BuildCmd:     "npm install",
		Artifact:     ".",
		Ports:        []int{3000},
		Env:          map[string]string{"NODE_ENV": "production"},
		Health:       "/health",
		Dependencies: []string{"express"},
	}

	// Check if using yarn
	for _, s := range snips {
		if s.Path == "package.json" && strings.Contains(s.Content, "yarn") {
			f.BuildTool = "yarn"
			f.BuildCmd = "yarn install"
			break
		}
	}

	return f, nil
}

// Dockerfile generates a Dockerfile for the project.
func (r *Rule) Dockerfile(ctx context.Context, f facts.Facts, c *llm.Client) (string, error) {
	// Convert facts to map for LLM
	factsMap := f.ToMap()

	// Generate Dockerfile
	dockerfile, err := c.GenerateDockerfile(ctx, factsMap)
	if err != nil {
		return "", fmt.Errorf("generate dockerfile: %w", err)
	}

	if os.Getenv("DG_DEBUG") == "1" {
		r.Log().Debug("generated dockerfile", "content", dockerfile)
	}

	return dockerfile, nil
}

// packageJSON represents the structure of a package.json file
type packageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// Detect returns a Node.js rule if the repository contains Node.js files.
func Detect(repo string) (*rules.StackRule, error) {
	// Look for package.json
	matches, err := doublestar.Glob(os.DirFS(repo), "**/package.json")
	if err != nil {
		return nil, fmt.Errorf("glob package.json: %w", err)
	}
	if len(matches) == 0 {
		return nil, nil
	}

	// Read the first package.json found
	pkgPath := filepath.Join(repo, matches[0])
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("read package.json: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	// Determine framework
	framework := "nodejs"
	if pkg.Dependencies != nil {
		switch {
		case pkg.Dependencies["react"] != "":
			framework = "react"
		case pkg.Dependencies["vue"] != "":
			framework = "vue"
		case pkg.Dependencies["@angular/core"] != "":
			framework = "angular"
		}
	}

	// Look for framework-specific files
	var frameworkFiles []string
	switch framework {
	case "react":
		frameworkFiles = []string{"**/src/index.js", "**/src/App.js", "**/vite.config.js"}
	case "vue":
		frameworkFiles = []string{"**/src/main.js", "**/vite.config.js", "**/vue.config.js"}
	case "angular":
		frameworkFiles = []string{"**/src/main.ts", "**/angular.json"}
	default:
		frameworkFiles = []string{"**/src/index.js", "**/app.js", "**/server.js"}
	}

	// Create the rule
	rule := &rules.StackRule{
		Name: framework,
		Signatures: append([]string{
			"**/package.json",
		}, frameworkFiles...),
		ManifestGlobs: []string{
			"**/package.json",
		},
		CodeGlobs: frameworkFiles,
		MainRegex: "require|import",
		BuildHints: map[string]string{
			"builder":   "node:20",
			"build_dir": filepath.Dir(matches[0]), // Set build directory to package.json location
		},
	}

	return rule, nil
}

func init() {
	rules.RegisterDetector(Detect)
}
