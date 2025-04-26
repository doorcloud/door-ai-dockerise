package docker

import (
	"context"
)

// BuildOptions contains options for building a Docker image
type BuildOptions struct {
	Context    string
	Tags       []string
	Dockerfile string
}

// Driver handles Docker operations
type Driver interface {
	Build(ctx context.Context, dockerfilePath string, opts BuildOptions) error
	Push(ctx context.Context, image string) error
}
