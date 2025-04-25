package e2e

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/internal/loop"
	"github.com/stretchr/testify/assert"
)

func TestReactIntegration(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping integration test; set DG_E2E=1 to run")
	}

	// Use local fixture
	repo := filepath.Join("test", "e2e", "fixtures", "react-min")

	// Run the Dockerfile generation loop
	ctx := context.Background()
	fsys := os.DirFS(repo)
	client := newTestClient(t)

	dockerfile, err := loop.Run(ctx, fsys, client)
	assert.NoError(t, err)
	assert.NotEmpty(t, dockerfile)

	// Build and run the container
	containerID, err := buildAndRun(t, repo, dockerfile, []string{"80:80"})
	assert.NoError(t, err)
	defer cleanupContainer(t, containerID)

	// Wait for the container to be ready
	waitForHTTP(t, "http://localhost:80", 60*time.Second)

	// Verify the app is running
	resp, err := http.Get("http://localhost:80")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
