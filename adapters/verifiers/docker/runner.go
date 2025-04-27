package docker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/facts/spring"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// Runner implements health verification for Docker containers
type Runner struct {
	client docker.Driver
}

// NewRunner creates a new Docker runner
func NewRunner(client docker.Driver) *Runner {
	return &Runner{
		client: client,
	}
}

// Run executes a container and verifies its health
func (r *Runner) Run(ctx context.Context, image string, spec *spring.Spec, logs io.Writer) error {
	// Generate a random port for the container
	port, err := r.client.GetRandomPort()
	if err != nil {
		return fmt.Errorf("failed to get random port: %w", err)
	}

	// Start the container
	containerID, err := r.client.Run(ctx, image, port, spec.Ports[0])
	if err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}
	defer r.client.Stop(ctx, containerID)
	defer r.client.Remove(ctx, containerID)

	// Stream container ID
	fmt.Fprintf(logs, "docker run   │ %s\n", containerID)

	// Stream container logs
	go func() {
		if err := r.client.Logs(ctx, containerID, logs); err != nil {
			fmt.Fprintf(logs, "error streaming logs: %v\n", err)
		}
	}()

	// Poll health endpoint
	startTime := time.Now()
	healthURL := fmt.Sprintf("http://localhost:%d%s", port, spec.HealthEndpoint)
	client := &http.Client{Timeout: 5 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("health check timeout: %w", ctx.Err())
		default:
			resp, err := client.Get(healthURL)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					elapsed := time.Since(startTime)
					fmt.Fprintf(logs, "health       │ OK in %.0f s\n", elapsed.Seconds())
					return nil
				}
			}
			time.Sleep(time.Second)
		}
	}
}
