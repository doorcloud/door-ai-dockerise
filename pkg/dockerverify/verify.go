package dockerverify

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Verify builds the Dockerfile in a temporary directory and returns the result
func Verify(ctx context.Context, fsys fs.FS, dockerfile string, timeout time.Duration) (bool, string, error) {
	// Create a temporary directory for the build
	dir, err := os.MkdirTemp("", "dockergen-*")
	if err != nil {
		return false, "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return false, "", fmt.Errorf("write Dockerfile: %w", err)
	}

	// Create .dockerignore
	if err := createDockerignore(dir); err != nil {
		return false, "", fmt.Errorf("create .dockerignore: %w", err)
	}

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return false, "", fmt.Errorf("create Docker client: %w", err)
	}

	// Build the image
	buildCtx, err := os.Open(dir)
	if err != nil {
		return false, "", fmt.Errorf("open build context: %w", err)
	}
	defer buildCtx.Close()

	resp, err := cli.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		return false, "", fmt.Errorf("build image: %w", err)
	}
	defer resp.Body.Close()

	// Read build output
	var logs strings.Builder
	if _, err := io.Copy(&logs, resp.Body); err != nil {
		return false, "", fmt.Errorf("read build output: %w", err)
	}

	// Get the last 100 lines of logs
	logLines := strings.Split(logs.String(), "\n")
	start := len(logLines) - 100
	if start < 0 {
		start = 0
	}
	lastLogs := strings.Join(logLines[start:], "\n")

	// Check if build was successful
	if strings.Contains(logs.String(), "Successfully built") {
		return true, lastLogs, nil
	}

	return false, lastLogs, nil
}

// createDockerignore creates a .dockerignore file in the build context directory
func createDockerignore(dir string) error {
	ignoreContent := `.git
**/*.iml
.idea
docs
*.md
!mvnw
!.mvn/**
`
	return os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(ignoreContent), 0644)
}
