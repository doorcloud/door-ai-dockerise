package facts

import (
	"context"
	"os"
	"strconv"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
)

type Static struct{}

func NewStatic() *Static {
	return &Static{}
}

func (s *Static) Facts(ctx context.Context, stack core.StackInfo) ([]string, error) {
	// Create a temporary directory for the filesystem
	dir, err := os.MkdirTemp("", "facts-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// Get facts from the legacy provider
	internalFacts, err := facts.GetFacts(dir)
	if err != nil {
		return nil, err
	}

	// Convert Facts struct to string slice
	var result []string
	if internalFacts.Language != "" {
		result = append(result, "language:"+internalFacts.Language)
	}
	if internalFacts.Framework != "" {
		result = append(result, "framework:"+internalFacts.Framework)
	}
	if internalFacts.BuildTool != "" {
		result = append(result, "build_tool:"+internalFacts.BuildTool)
	}
	if internalFacts.BuildCmd != "" {
		result = append(result, "build_cmd:"+internalFacts.BuildCmd)
	}
	if internalFacts.BuildDir != "" {
		result = append(result, "build_dir:"+internalFacts.BuildDir)
	}
	if internalFacts.StartCmd != "" {
		result = append(result, "start_cmd:"+internalFacts.StartCmd)
	}
	if internalFacts.Artifact != "" {
		result = append(result, "artifact:"+internalFacts.Artifact)
	}
	if len(internalFacts.Ports) > 0 {
		for _, port := range internalFacts.Ports {
			result = append(result, "port:"+strconv.Itoa(port))
		}
	}
	if internalFacts.Health != "" {
		result = append(result, "health:"+internalFacts.Health)
	}
	if len(internalFacts.Env) > 0 {
		for k, v := range internalFacts.Env {
			result = append(result, "env:"+k+"="+v)
		}
	}
	if internalFacts.BaseImage != "" {
		result = append(result, "base_image:"+internalFacts.BaseImage)
	}
	if internalFacts.HasLockfile {
		result = append(result, "has_lockfile:true")
	}

	return result, nil
}
