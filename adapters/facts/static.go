package facts

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Static provides static facts about a project
type Static struct{}

// NewStatic creates a new static fact provider
func NewStatic() *Static {
	return &Static{}
}

// Gather implements the Provider interface
func (s *Static) Gather(projectDir string) (map[string]interface{}, error) {
	// Return some static facts
	return map[string]interface{}{
		"environment": "production",
		"platform":    "linux/amd64",
	}, nil
}

func (s *Static) Facts(ctx context.Context, stack core.StackInfo) ([]core.Fact, error) {
	// Create a temporary directory for the filesystem
	dir, err := os.MkdirTemp("", "facts-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// Get facts based on the stack type
	switch stack.Name {
	case "react":
		return []core.Fact{
			{Key: "framework", Value: "react"},
			{Key: "language", Value: "javascript"},
			{Key: "buildTool", Value: "npm"},
		}, nil
	case "node":
		return []core.Fact{
			{Key: "runtime", Value: "node"},
			{Key: "language", Value: "javascript"},
			{Key: "buildTool", Value: "npm"},
		}, nil
	case "spring-boot":
		return []core.Fact{
			{Key: "framework", Value: "spring-boot"},
			{Key: "language", Value: "java"},
			{Key: "buildTool", Value: "maven"},
		}, nil
	default:
		return nil, nil
	}
}
