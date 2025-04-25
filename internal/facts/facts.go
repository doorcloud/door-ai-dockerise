package facts

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

//go:embed prompts/facts.tmpl
var promptsFS embed.FS

// Facts represents the detected facts about a technology stack
type Facts struct {
	Language  string            `json:"language"`   // "java", "node", "python"…
	Framework string            `json:"framework"`  // "spring-boot", "express", "flask"…
	BuildTool string            `json:"build_tool"` // "maven", "npm", "pip", …
	BuildCmd  string            `json:"build_cmd"`  // e.g. "mvn package", "npm run build"
	BuildDir  string            `json:"build_dir"`  // directory containing build files
	StartCmd  string            `json:"start_cmd"`  // e.g. "java -jar app.jar"
	Artifact  string            `json:"artifact"`   // glob or relative path
	Ports     []int             `json:"ports"`      // e.g. [8080], [3000]
	Health    string            `json:"health"`     // URL path or CMD
	Env       map[string]string `json:"env"`        // e.g. {"NODE_ENV": "production"}
	BaseImage string            `json:"base_image"` // e.g. "eclipse-temurin:17-jdk"
}

// Infer analyzes the filesystem and returns facts about the project
func Infer(ctx context.Context, fsys fs.FS, rule detect.Rule) (Facts, error) {
	client := llm.New()
	return InferWithClient(ctx, fsys, rule, client)
}

// InferWithClient analyzes the project using the provided client
func InferWithClient(ctx context.Context, fsys fs.FS, rule detect.Rule, client llm.Client) (Facts, error) {
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
