package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliou/dockerfile-gen/internal/loop"
)

func TestE2E_SpringBoot(t *testing.T) {
	// Skip if running in CI without E2E flag
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test; set DG_E2E=1 to run")
	}

	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	// Get the Spring Boot test project
	testDir := filepath.Join("testdata", "springboot")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("test directory %s does not exist", testDir)
	}

	// Run through the loop package directly
	t.Run("via loop package", func(t *testing.T) {
		fsys := os.DirFS(testDir)
		dockerfile, err := loop.Run(context.Background(), fsys)
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
	cmd := exec.Command("go", "build", "-o", tmpfile.Name(), "../../cmd/dockergen")
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
