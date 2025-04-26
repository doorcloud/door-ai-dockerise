package core

import (
	"context"
	"io"
)

// BuildInput contains the necessary data for building a Docker image
type BuildInput struct {
	ContextTar io.Reader // .tar.gz of source tree
	Dockerfile string    // Dockerfile text
}

// ImageRef represents a built Docker image
type ImageRef struct {
	Name string // e.g. doorai/gen:sha256-abcdef
}

// String returns the image name
func (r ImageRef) String() string {
	return r.Name
}

// BuildDriver defines the interface for building Docker images
type BuildDriver interface {
	// Build creates a Docker image from the given input
	// The out stream receives live logs; Orchestrator forwards it to CLI / API
	Build(ctx context.Context, in BuildInput, out io.Writer) (ImageRef, error)
}
