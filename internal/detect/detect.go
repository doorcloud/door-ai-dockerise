package detect

import (
	"errors"
	"io/fs"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/registry"
)

// ErrUnknownStack is returned when no rule matches the project
var ErrUnknownStack = errors.New("unknown technology stack")

// RuleInfo represents information about a detected technology stack
type RuleInfo struct {
	Name string // e.g. "spring-boot"
	Tool string // e.g. "maven"
}

// Detect checks if the given filesystem matches any known rules
func Detect(fsys fs.FS) (RuleInfo, error) {
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
			return RuleInfo{
				Name: rule.Name(),
				Tool: tool,
			}, nil
		}
	}

	return RuleInfo{}, ErrUnknownStack
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
