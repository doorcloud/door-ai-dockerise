package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/doorcloud/door-ai-dockerise/adapters/log/plain"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/logs"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/mock"
)

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = logs.NewTagWriter(os.Stdout, "cmd")
	cmd.Stderr = logs.NewTagWriter(os.Stderr, "cmd")
	return cmd.Run()
}

func newTestClient(t *testing.T) *mock.Client {
	return mock.New(map[string]string{
		"stack:react\nframework:react": `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`,
	})
}

func buildAndRun(t *testing.T, dir string, dockerfile string, ports []string) (string, error) {
	// Write Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return "", fmt.Errorf("write Dockerfile: %w", err)
	}

	// Build image
	buildCtx := filepath.Join(dir, ".")
	cmd := exec.Command("docker", "build", "-t", "test-image", buildCtx)
	cmd.Stdout = logs.NewTagWriter(os.Stdout, "docker build")
	cmd.Stderr = logs.NewTagWriter(os.Stderr, "docker build")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build image: %w", err)
	}

	// Run container
	args := []string{"run", "-d"}
	for _, p := range ports {
		args = append(args, "-p", p)
	}
	args = append(args, "test-image")
	cmd = exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run container: %w", err)
	}

	logs.WriteTagged(os.Stdout, "docker run", "%s", output)
	return string(output), nil
}

func cleanupContainer(t *testing.T, containerID string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.43"))
	if err != nil {
		t.Logf("Failed to create Docker client: %v", err)
		return
	}

	timeoutSeconds := 10
	err = cli.ContainerStop(context.Background(), containerID, container.StopOptions{Timeout: &timeoutSeconds})
	if err != nil {
		t.Logf("Failed to stop container: %v", err)
	}

	err = cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	if err != nil {
		t.Logf("Failed to remove container: %v", err)
	}
}

func waitForHTTP(t *testing.T, url string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("HTTP endpoint not ready after %s", timeout)
}

// NewRecorder creates a pipe where the writer end is wrapped in a LogStreamer
// and the reader end can be used to verify the logs.
func NewRecorder() (io.Reader, core.LogStreamer) {
	pr, pw := io.Pipe()
	streamer := plain.NewWriterStreamer(pw)
	return pr, streamer
}
