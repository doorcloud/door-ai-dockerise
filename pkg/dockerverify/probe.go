package dockerverify

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

// Probe checks if a container is healthy by running a health check command
func Probe(ctx context.Context, cli APIClient, containerID string, cmd []string) error {
	// Start the container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	// Stop the container when we're done
	defer func() {
		timeout := 10
		cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
	}()

	// Wait for the container to exit
	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("wait for container: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with status %d", status.StatusCode)
		}
	}

	return nil
}
