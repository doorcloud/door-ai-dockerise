package facts

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

//go:embed prompts/facts.tmpl
var promptsFS embed.FS

// Facts represents information about a project
type Facts struct {
	Language  string            `json:"language"`
	Framework string            `json:"framework"`
	BuildTool string            `json:"build_tool"`
	BuildCmd  string            `json:"build_cmd"`
	BuildDir  string            `json:"build_dir"`
	StartCmd  string            `json:"start_cmd"`
	Artifact  string            `json:"artifact"`
	Ports     []int             `json:"ports"`
	Health    string            `json:"health"`
	BaseImage string            `json:"base_image"`
	Env       map[string]string `json:"env"`
}

// GetFacts extracts facts about the project in the given directory
func GetFacts(dir string) (Facts, error) {
	fsys := os.DirFS(dir)
	rule, err := detect.Detect(fsys)
	if err != nil {
		return Facts{}, err
	}
	return GetFactsFromRule(fsys, rule)
}

// GetFactsFromRule extracts facts about the project using the given rule
func GetFactsFromRule(fsys fs.FS, rule detect.RuleInfo) (Facts, error) {
	switch rule.Name {
	case "spring-boot":
		return getSpringBootFacts(fsys, rule)
	case "node":
		return getNodeFacts(fsys, rule)
	default:
		return Facts{}, nil
	}
}

func getSpringBootFacts(fsys fs.FS, rule detect.RuleInfo) (Facts, error) {
	return Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: rule.Tool,
		BuildCmd:  "./mvnw -q package -DskipTests",
		BuildDir:  ".",
		StartCmd:  "java -jar target/*.jar",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Health:    "/actuator/health",
		BaseImage: "openjdk:11-jdk",
		Env:       map[string]string{},
	}, nil
}

func getNodeFacts(fsys fs.FS, rule detect.RuleInfo) (Facts, error) {
	return Facts{
		Language:  "javascript",
		Framework: "node",
		BuildTool: rule.Tool,
		BuildCmd:  "npm install",
		BuildDir:  ".",
		StartCmd:  "npm start",
		Artifact:  ".",
		Ports:     []int{3000},
		Health:    "/health",
		BaseImage: "node:18-alpine",
		Env:       map[string]string{},
	}, nil
}

// Infer analyzes the filesystem and returns facts about the project
func Infer(ctx context.Context, fsys fs.FS, rule detect.RuleInfo) (Facts, error) {
	client := llm.New()
	return InferWithClient(ctx, fsys, rule, client)
}

// InferWithClient analyzes the project using the provided client
func InferWithClient(ctx context.Context, fsys fs.FS, rule detect.RuleInfo, client llm.Client) (Facts, error) {
	// Get relevant snippets
	snippets, err := getSnippets(fsys)
	if err != nil {
		return Facts{}, fmt.Errorf("get snippets: %w", err)
	}

	// Build prompt from template
	prompt, err := buildFactsPrompt(snippets)
	if err != nil {
		return Facts{}, fmt.Errorf("build prompt: %w", err)
	}

	// Get facts from LLM
	response, err := client.Chat(prompt, "facts")
	if err != nil {
		return Facts{}, fmt.Errorf("chat: %w", err)
	}

	// Parse response
	var facts Facts
	if err := json.Unmarshal([]byte(response), &facts); err != nil {
		return Facts{}, fmt.Errorf("parse response: %w", err)
	}

	return facts, nil
}

// getSnippets extracts relevant code snippets from the filesystem
func getSnippets(fsys fs.FS) ([]string, error) {
	var snippets []string

	// Walk the filesystem
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip non-code files
		if !isCodeFile(path) {
			return nil
		}

		// Read file content
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		// Add to snippets
		snippets = append(snippets, string(content))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return snippets, nil
}

// isCodeFile checks if a file is likely to contain code
func isCodeFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".java" || ext == ".py" || ext == ".js" || ext == ".ts" || ext == ".go"
}

// buildFactsPrompt creates the prompt for fact extraction
func buildFactsPrompt(snippets []string) (string, error) {
	// Read template
	tmplContent, err := promptsFS.ReadFile("prompts/facts.tmpl")
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("facts").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Execute template
	var result strings.Builder
	data := struct {
		Snippets string
	}{
		Snippets: strings.Join(snippets, "\n\n"),
	}
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return result.String(), nil
}
