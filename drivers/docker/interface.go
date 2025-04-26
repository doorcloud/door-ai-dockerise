package docker

import (
	"context"
	"io"
)

// BuildOptions contains options for building a Docker image
type BuildOptions struct {
	Tags       []string
	Dockerfile string
	BuildArgs  map[string]string
}

// Driver defines the interface for Docker operations
type Driver interface {
	Build(ctx context.Context, context io.Reader, options BuildOptions) error
	Push(ctx context.Context, image string) error
}
