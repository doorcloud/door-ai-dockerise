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
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

//go:embed prompts/facts.tmpl
var promptsFS embed.FS

// GetFacts extracts facts about the project in the given directory
func GetFacts(dir string) (types.Facts, error) {
	fsys := os.DirFS(dir)
	reg := rules.NewRegistry()
	rule, ok := reg.Detect(fsys)
	if !ok {
		return types.Facts{}, detect.ErrUnknownStack
	}
	return GetFactsFromRule(fsys, rule)
}

// GetFactsFromRule extracts facts about the project using the given rule
func GetFactsFromRule(fsys fs.FS, rule detect.RuleInfo) (types.Facts, error) {
	return rules.GetFacts(fsys, rule)
}

// Infer analyzes the filesystem and returns facts about the project
func Infer(ctx context.Context, fsys fs.FS, rule detect.RuleInfo) (types.Facts, error) {
	client := llm.New()
	return InferWithClient(ctx, fsys, rule, client)
}

// InferWithClient analyzes the project using the provided client
func InferWithClient(ctx context.Context, fsys fs.FS, rule detect.RuleInfo, client llm.Client) (types.Facts, error) {
	// Get relevant snippets
	snippets, err := getSnippets(fsys)
	if err != nil {
		return types.Facts{}, fmt.Errorf("get snippets: %w", err)
	}

	// Build prompt from template
	prompt, err := buildFactsPrompt(snippets, rule.Tool)
	if err != nil {
		return types.Facts{}, fmt.Errorf("build prompt: %w", err)
	}

	// Get facts from LLM
	response, err := client.Chat(prompt, "facts")
	if err != nil {
		return types.Facts{}, fmt.Errorf("chat: %w", err)
	}

	// Parse response
	var facts types.Facts
	if err := json.Unmarshal([]byte(response), &facts); err != nil {
		return types.Facts{}, fmt.Errorf("parse response: %w", err)
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
func buildFactsPrompt(snippets []string, tool string) (string, error) {
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
		Tool     string
	}{
		Snippets: strings.Join(snippets, "\n\n"),
		Tool:     tool,
	}
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return result.String(), nil
}
