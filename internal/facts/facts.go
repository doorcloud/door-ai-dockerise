package facts

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

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
	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return Facts{}, fmt.Errorf("OPENAI_API_KEY is required")
	}
	client := llm.NewClient(apiKey)

	return InferWithClient(ctx, fsys, rule, client)
}

// InferWithClient analyzes the project using the provided client
func InferWithClient(ctx context.Context, fsys fs.FS, rule detect.Rule, client llm.Client) (Facts, error) {
	// Get relevant snippets
	snippets, err := getSnippets(fsys)
	if err != nil {
		return Facts{}, fmt.Errorf("get snippets: %w", err)
	}

	// Build facts prompt
	prompt := buildFactsPrompt(snippets)

	// Call LLM to analyze facts
	resp, err := client.Chat(ctx, prompt)
	if err != nil {
		return Facts{}, fmt.Errorf("llm call failed: %w", err)
	}

	// Parse response as JSON
	var facts Facts
	if err := json.Unmarshal([]byte(resp), &facts); err != nil {
		return Facts{}, fmt.Errorf("parse facts json: %w", err)
	}

	return facts, nil
}

// getSnippets collects relevant code snippets from the filesystem
func getSnippets(fsys fs.FS) ([]string, error) {
	var snippets []string

	// Collect relevant files
	files := []string{
		"pom.xml",
		"build.gradle",
		"application.properties",
		"application.yml",
		"application.yaml",
	}

	for _, file := range files {
		content, err := fs.ReadFile(fsys, file)
		if err == nil {
			snippets = append(snippets, string(content))
		}
	}

	// Find and add main application class
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (filepath.Ext(path) == ".java" || filepath.Ext(path) == ".kt") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			if regexp.MustCompile(`@SpringBootApplication`).Match(content) {
				snippets = append(snippets, string(content))
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return snippets, nil
}

// buildFactsPrompt creates the prompt for fact extraction
func buildFactsPrompt(snippets []string) string {
	return fmt.Sprintf(`You are a code analysis expert. Given a set of code snippets, extract key facts about the project.
The output must be valid JSON with the following structure:
{
  "language": "java",
  "framework": "spring-boot",
  "build_tool": "build system (maven, gradle, etc)",
  "build_cmd": "command to build",
  "build_dir": "directory containing build files (e.g., '.', 'backend/')",
  "start_cmd": "command to start the application",
  "artifact": "path to built artifact",
  "ports": [8080],
  "env": {"key": "value"},
  "health": "/actuator/health"
}

Code snippets:
%s`, snippets)
}
