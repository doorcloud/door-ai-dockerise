package generate

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Generator defines the interface for generating Dockerfiles
type Generator interface {
	// Generate creates a Dockerfile based on the provided facts
	Generate(ctx context.Context, facts core.Facts) (string, error)
}
