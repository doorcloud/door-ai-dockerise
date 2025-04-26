package detect

import (
	"errors"
	"io/fs"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/registry"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// ErrUnknownStack is returned when no rule matches the project
var ErrUnknownStack = errors.New("unknown technology stack")

// Detect checks if the given filesystem matches any known rules
func Detect(fsys fs.FS) (types.RuleInfo, error) {
	// Try all registered rules
	for _, rule := range registry.All() {
		if rule.Detect(fsys) {
			facts := rule.Facts(fsys)
			tool := "npm" // default to npm
			if facts != nil {
				if buildTool, ok := facts["build_tool"].(string); ok {
					tool = buildTool
				}
			}
			return types.RuleInfo{
				Name: rule.Name(),
				Tool: tool,
			}, nil
		}
	}

	return types.RuleInfo{}, ErrUnknownStack
}

// DetectStack analyzes the given directory and returns the detected technology stack
func DetectStack(dir string) (string, error) {
	fsys := os.DirFS(dir)
	rule, err := Detect(fsys)
	if err != nil {
		return "", err
	}
	return rule.Name, nil
}

// DetectProjectType detects the project type using the provided registry
func DetectProjectType(fsys fs.FS, registry interface{}) (string, error) {
	rule, err := Detect(fsys)
	if err != nil {
		return "", err
	}
	return rule.Name, nil
}
