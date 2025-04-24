//go:build integration
// +build integration

package e2e

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/pkg/dockerverify"
)

func TestE2E_SpringBootRepos(t *testing.T) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.E2E {
		t.Skip("Skipping E2E test. Set DG_E2E=1 to run.")
	}

	if cfg.OpenAIKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	t.Log("Starting Spring Boot integration test...")

	// Create LLM client
	llmClient, err := llm.NewClient(&cfg)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}
	t.Log("LLM client created successfully")

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"),
	)
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	t.Log("Docker client created successfully")

	// Test cases
	testCases := []struct {
		name     string
		repo     string
		expected facts.Facts
	}{
		{
			name: "Spring Boot",
			repo: "https://github.com/spring-projects/spring-petclinic.git",
			expected: facts.Facts{
				Language:  "java",
				Framework: "spring-boot",
				BuildTool: "maven",
				BuildDir:  ".",
				Ports:     []int{8080},
				Health:    "/actuator/health",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Starting test case: %s", tc.name)
			t.Logf("Cloning repository: %s", tc.repo)

			// Create temporary directory for the repository
			tempDir, err := os.MkdirTemp("", "dockergen-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)
			t.Logf("Created temporary directory: %s", tempDir)

			// Clone repository
			startTime := time.Now()
			cmd := exec.Command("git", "clone", tc.repo, tempDir)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to clone repository: %v\n%s", err, output)
			}
			t.Logf("Repository cloned in %v", time.Since(startTime))

			// Verify Dockerfile with retries
			ctx := context.Background()
			t.Log("Starting Dockerfile verification...")
			startTime = time.Now()
			dockerfile, err := dockerverify.VerifyDockerfile(ctx, dockerClient, tempDir, tc.expected, llmClient, 4, &cfg)
			if err != nil {
				t.Fatalf("Failed to verify Dockerfile: %v", err)
			}
			t.Logf("Dockerfile verification completed in %v", time.Since(startTime))

			// Log the generated Dockerfile
			t.Logf("Generated Dockerfile:\n%s", dockerfile)
		})
	}
}

func captureLogOutput() string {
	// Create a pipe to capture log output
	r, w, err := os.Pipe()
	if err != nil {
		return ""
	}

	// Save original stderr
	originalStderr := os.Stderr
	os.Stderr = w

	// Create a channel to signal when we're done
	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	// Restore stderr
	os.Stderr = originalStderr
	w.Close()

	// Wait for the copy to complete
	return <-done
}
