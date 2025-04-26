package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

func runCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func newTestClient(t *testing.T) llm.Client {
	if os.Getenv("DG_MOCK_LLM") != "" {
		return &llm.MockClient{}
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Fatal("OPENAI_API_KEY environment variable is not set")
	}
	return llm.New()
}

func buildAndRun(t *testing.T, dir string, dockerfile string, ports []string) (string, error) {
	// Write Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return "", fmt.Errorf("write Dockerfile: %w", err)
	}

	// Build image
	buildCtx := filepath.Join(dir, ".")
	cmd := exec.Command("docker", "build", "-t", "test-image", buildCtx)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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

	err = cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
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
