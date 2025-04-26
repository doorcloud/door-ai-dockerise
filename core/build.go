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

// LogStreamer defines the interface for streaming build logs
type LogStreamer interface {
	// Info writes an informational message
	Info(msg string)
	// Error writes an error message
	Error(msg string)
	// Write implements io.Writer for raw log output
	Write(p []byte) (n int, err error)
}

// BuildDriver defines the interface for building Docker images
type BuildDriver interface {
	// Build creates a Docker image from the given input
	// The log streamer receives live logs; Orchestrator forwards it to CLI / API
	Build(ctx context.Context, in BuildInput, log LogStreamer) (ImageRef, error)
}
