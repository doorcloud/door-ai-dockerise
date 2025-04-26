package generate

import (
	"context"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/dockerfile"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

type LLM struct {
	client llm.Client
}

func NewLLM(client llm.Client) *LLM {
	return &LLM{
		client: client,
	}
}

func (l *LLM) Generate(ctx context.Context, facts []string) (string, error) {
	// Convert string facts to types.Facts
	internalFacts := types.Facts{}
	for _, fact := range facts {
		// Parse fact string into key-value pairs
		// This is a simplified version - you might need more complex parsing
		if len(fact) > 0 {
			parts := strings.SplitN(fact, ":", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				switch key {
				case "language":
					internalFacts.Language = value
				case "framework":
					internalFacts.Framework = value
				case "build_tool":
					internalFacts.BuildTool = value
				case "build_cmd":
					internalFacts.BuildCmd = value
				case "build_dir":
					internalFacts.BuildDir = value
				case "start_cmd":
					internalFacts.StartCmd = value
				case "artifact":
					internalFacts.Artifact = value
				case "health":
					internalFacts.Health = value
				case "base_image":
					internalFacts.BaseImage = value
				case "has_lockfile":
					internalFacts.HasLockfile = value == "true"
				}
			}
		}
	}

	return dockerfile.Generate(ctx, internalFacts, l.client)
}
