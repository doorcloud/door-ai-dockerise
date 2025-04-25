//go:build integration

package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/internal/loop"
)

// TestIntegration_SpringBoot tests the full pipeline with a Spring Boot project
func TestIntegration_SpringBoot(t *testing.T) {
	// Skip if running without E2E flag
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test; set DG_E2E=1 to run")
	}

	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	// Skip if OpenAI key not set
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Set build timeout from env or default
	timeout := 20 * time.Minute
	if val := os.Getenv("GO_BUILD_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			timeout = d
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get the Spring Boot test project
	testDir := filepath.Join("testdata", "springboot")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("test directory %s does not exist", testDir)
	}

	// Run through the loop package directly
	t.Run("via loop package", func(t *testing.T) {
		fsys := os.DirFS(testDir)
		dockerfile, err := loop.Run(ctx, fsys)
		if err != nil {
			t.Fatalf("loop.Run failed: %v", err)
		}

		verifyDockerfile(t, dockerfile)
	})

	// Run through the CLI
	t.Run("via CLI", func(t *testing.T) {
		// Build the CLI
		cliPath := buildCLI(t)
		defer os.Remove(cliPath)

		// Run the CLI
		cmd := exec.Command(cliPath, testDir)
		cmd.Env = append(os.Environ(),
			"GO_BUILD_TIMEOUT="+timeout.String(),
			"DG_E2E=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
		}

		verifyDockerfile(t, string(output))
	})
}

func buildCLI(t *testing.T) string {
	t.Helper()

	// Create temporary file for the binary
	tmpfile, err := os.CreateTemp("", "dockergen-test-")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpfile.Close()

	// Build the CLI
	cmd := exec.Command("go", "build", "-tags=integration", "-o", tmpfile.Name(), "../../cmd/dockergen")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build CLI: %v\nOutput: %s", err, output)
	}

	return tmpfile.Name()
}

func verifyDockerfile(t *testing.T, dockerfile string) {
	t.Helper()

	// Check required commands
	requiredCommands := []string{
		"FROM",
		"WORKDIR",
		"COPY",
		"RUN",
		"EXPOSE",
		"HEALTHCHECK",
		"CMD",
	}

	for _, cmd := range requiredCommands {
		if !strings.Contains(dockerfile, cmd) {
			t.Errorf("Dockerfile missing command: %s", cmd)
		}
	}

	// Check Spring Boot specific elements
	springBootElements := []string{
		"eclipse-temurin:17-jdk",
		"./mvnw -q package",
		"target/*.jar",
		"8080",
		"/actuator/health",
	}

	for _, element := range springBootElements {
		if !strings.Contains(dockerfile, element) {
			t.Errorf("Dockerfile missing Spring Boot element: %s", element)
		}
	}
}
