package dockerverify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// Probe probes a container to check if it's healthy
func Probe(ctx context.Context, containerID string, cli DockerClient, healthEndpoint string, timeout time.Duration) (bool, error) {
	// Start the container
	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return false, fmt.Errorf("start container: %w", err)
	}
	defer cli.ContainerStop(ctx, containerID, container.StopOptions{})

	// Get container info
	info, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("inspect container: %w", err)
	}

	// Wait for container to be ready
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Check if container is running
		info, err := cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return false, fmt.Errorf("inspect container: %w", err)
		}
		if info.State.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Get container IP
	ip := info.NetworkSettings.IPAddress
	if ip == "" {
		return false, fmt.Errorf("container has no IP address")
	}

	// Probe the health endpoint
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d%s", ip, 8080, healthEndpoint))
	if err != nil {
		return false, fmt.Errorf("probe health endpoint: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read response body: %w", err)
	}

	// Check if response indicates healthy status
	return strings.Contains(string(body), `"status":"UP"`), nil
}
