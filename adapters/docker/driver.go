package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

// Driver handles Docker operations
type Driver struct {
	client *client.Client
}

// NewDriver creates a new Docker driver
func NewDriver() (*Driver, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Driver{
		client: cli,
	}, nil
}

// BuildImage builds a Docker image from a Dockerfile
func (d *Driver) BuildImage(ctx context.Context, dockerfilePath string, tag string) error {
	// Create a build context
	buildCtx, err := archive.TarWithOptions(filepath.Dir(dockerfilePath), &archive.TarOptions{
		IncludeFiles: []string{filepath.Base(dockerfilePath)},
	})
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildCtx.Close()

	// Build the image
	buildResp, err := d.client.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile: filepath.Base(dockerfilePath),
		Tags:       []string{tag},
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer buildResp.Body.Close()

	// Read the build output
	_, err = io.Copy(os.Stdout, buildResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read build output: %w", err)
	}

	return nil
}

// RunContainer runs a container from an image
func (d *Driver) RunContainer(ctx context.Context, image string) error {
	// Create the container
	resp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   []string{"echo", "Container is running"},
	}, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for the container to finish
	statusCh, errCh := d.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case <-statusCh:
	}

	// Remove the container
	if err := d.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// VerifyDockerfile builds and runs a container to verify the Dockerfile
func (d *Driver) VerifyDockerfile(ctx context.Context, dockerfilePath string) error {
	// Build the image
	imageTag := "test-image"
	if err := d.BuildImage(ctx, dockerfilePath, imageTag); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Run the container
	if err := d.RunContainer(ctx, imageTag); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	return nil
}
