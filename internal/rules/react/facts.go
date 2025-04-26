package react

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

// FactsDetector extracts the static bits we need for the LLM.
type FactsDetector struct{}

func (FactsDetector) Name() string {
	return "react"
}

func (FactsDetector) Detect(fsys fs.FS) bool {
	return (&ReactDetector{}).Detect(fsys)
}

func (FactsDetector) Facts(fsys fs.FS) map[string]any {
	return (&ReactDetector{}).Facts(fsys)
}

func (FactsDetector) InferFacts(ctx context.Context, cli llm.Client, deps map[string]string) (map[string]any, error) {
	prompt := fmt.Sprintf(`Given these npm dependencies:
%s

Extract the following facts about this React application:
1. build_tool: npm or yarn
2. build_cmd: the command to build the static site
3. ports: array of ports to expose (default [80])
4. health: health check endpoint (default "/")

Return a JSON object with these fields.`, formatDeps(deps))

	resp, err := cli.Chat(prompt, "facts")
	if err != nil {
		return nil, fmt.Errorf("infer facts: %w", err)
	}

	var facts map[string]any
	if err := json.Unmarshal([]byte(resp), &facts); err != nil {
		return nil, fmt.Errorf("parse facts: %w", err)
	}

	// Ensure required fields exist
	if _, ok := facts["build_tool"]; !ok {
		facts["build_tool"] = "npm"
	}
	if _, ok := facts["build_cmd"]; !ok {
		facts["build_cmd"] = "npm ci && npm run build"
	}
	if _, ok := facts["ports"]; !ok {
		facts["ports"] = []int{80}
	}
	if _, ok := facts["health"]; !ok {
		facts["health"] = "/"
	}

	return facts, nil
}

func formatDeps(deps map[string]string) string {
	var s string
	for name, version := range deps {
		s += fmt.Sprintf("- %s@%s\n", name, version)
	}
	return s
}
