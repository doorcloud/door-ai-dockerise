package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// Engine implements the core.BuildDriver interface using the local Docker socket
type Engine struct {
	cli *client.Client
}

// NewEngine creates a new Docker engine client
func NewEngine() (*Engine, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"), // Set API version to match server
	)
	if err != nil {
		return nil, err
	}
	return &Engine{cli: cli}, nil
}

// Build implements the core.BuildDriver interface
func (e *Engine) Build(ctx context.Context, in core.BuildInput, log core.LogStreamer) (core.ImageRef, error) {
	// Set build options
	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"doorai/gen:latest"},
		Remove:     true,
	}

	// Start the build
	resp, err := e.cli.ImageBuild(ctx, in.ContextTar, opts)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to start build: %v", err))
		return core.ImageRef{}, err
	}
	defer resp.Body.Close()

	// Stream build output
	_, err = io.Copy(log, resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to stream build output: %v", err))
		return core.ImageRef{}, err
	}

	// Return the image reference
	return core.ImageRef{Name: "doorai/gen:latest"}, nil
}
