package facts

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type Static struct{}

func NewStatic() *Static {
	return &Static{}
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
	case "springboot":
		return []core.Fact{
			{Key: "framework", Value: "springboot"},
			{Key: "language", Value: "java"},
			{Key: "buildTool", Value: "maven"},
		}, nil
	default:
		return nil, nil
	}
}
