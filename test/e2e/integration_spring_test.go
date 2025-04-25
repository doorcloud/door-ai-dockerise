package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aliou/dockerfile-gen/internal/config"
	"github.com/aliou/dockerfile-gen/internal/generate"
	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/stretchr/testify/assert"
)

func TestE2E_SpringBootRepos(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping e2e test; set DG_E2E=1 to run")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test; OPENAI_API_KEY not set")
	}

	// Create LLM client
	client, err := llm.NewOpenAIClient(apiKey)
	assert.NoError(t, err)

	// Create config
	cfg := config.New()

	// Test cases
	testCases := []struct {
		name string
		repo string
	}{
		{
			name: "spring-petclinic",
			repo: "https://github.com/spring-projects/spring-petclinic",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			// Clone repository
			dir := t.TempDir()
			err := cloneRepo(ctx, tc.repo, dir)
			assert.NoError(t, err)

			// Generate Dockerfile
			dockerfile, err := generate.Generate(ctx, os.DirFS(dir), client, 3, cfg.BuildTimeout)
			assert.NoError(t, err)
			assert.NotEmpty(t, dockerfile)
		})
	}
}

func cloneRepo(ctx context.Context, url string, dir string) error {
	// TODO: Implement repository cloning
	return nil
}

func TestSpringBootIntegration(t *testing.T) {
	// Skip if OPENAI_API_KEY is not set
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Create a new LLM client
	cli, err := llm.NewClient()
	if err != nil {
		t.Fatalf("failed to create LLM client: %v", err)
	}

	// Get the test data directory
	testDir := filepath.Join("testdata", "spring")
	fsys := os.DirFS(testDir)

	// Generate Dockerfile
	dockerfile, err := generate.Generate(context.Background(), fsys, cli, 3, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to generate Dockerfile: %v", err)
	}

	// Verify the Dockerfile contains expected commands
	expectedCommands := []string{
		"FROM",
		"WORKDIR",
		"COPY",
		"RUN",
		"CMD",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(dockerfile, cmd) {
			t.Errorf("Dockerfile missing command: %s", cmd)
		}
	}
}
