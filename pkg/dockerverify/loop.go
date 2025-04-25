package dockerverify

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// VerifyLoop verifies a Dockerfile by building and probing it in a loop
func VerifyLoop(ctx context.Context, fsys fs.FS, dockerfile string, healthEndpoint string, maxRetries int, buildTimeout time.Duration) (bool, string, error) {
	// Initialize Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"),
	)
	if err != nil {
		return false, "", fmt.Errorf("create Docker client: %w", err)
	}

	// Build the image
	buildLogs, err := Build(ctx, fsys, dockerfile, cli)
	if err != nil {
		return false, buildLogs, fmt.Errorf("build failed: %w", err)
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "dockergen-test",
	}, nil, nil, &v1.Platform{}, "")
	if err != nil {
		return false, buildLogs, fmt.Errorf("create container: %w", err)
	}
	defer cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

	// Probe the container
	healthy, err := Probe(ctx, resp.ID, cli, healthEndpoint, buildTimeout)
	if err != nil {
		return false, buildLogs, fmt.Errorf("probe failed: %w", err)
	}

	if !healthy {
		return false, buildLogs, fmt.Errorf("container not healthy")
	}

	return true, buildLogs, nil
}
