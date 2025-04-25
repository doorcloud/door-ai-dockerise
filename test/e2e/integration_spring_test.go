package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegration_Spring(t *testing.T) {
	// Skip if running without E2E flag
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test; set DG_E2E=1 to run")
	}

	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	// Set build timeout from env or default
	timeout := 20 * time.Minute
	if val := os.Getenv("GO_BUILD_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			timeout = d
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get the Spring Boot test project
	testDir := filepath.Join("testdata", "springboot")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("test directory %s does not exist", testDir)
	}

	// Run Docker build with context
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", "test-spring", "-f", "Dockerfile", testDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("docker build failed: %v\nOutput: %s", err, out)
	}
}
