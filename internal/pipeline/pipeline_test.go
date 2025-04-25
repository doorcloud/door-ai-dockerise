package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	_ "github.com/doorcloud/door-ai-dockerise/internal/rules/springboot"
)

func TestRun(t *testing.T) {
	// Skip if Docker is not available
	if os.Getenv("SKIP_DOCKER") == "1" {
		t.Skip("Skipping test that requires Docker")
	}

	// Set up test environment
	os.Setenv("DG_MOCK_LLM", "1")
	defer os.Unsetenv("DG_MOCK_LLM")

	// Use test directory with pom.xml
	testDir := filepath.Join("testdata")

	client := llm.New()

	if err := Run(testDir, client); err != nil {
		t.Errorf("Run() error = %v", err)
		return
	}

	// Check that Dockerfile was created
	df := filepath.Join(testDir, "Dockerfile")
	if _, err := os.Stat(df); err != nil {
		t.Errorf("Run() did not create Dockerfile: %v", err)
	}
}
