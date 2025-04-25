package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	_ "github.com/doorcloud/door-ai-dockerise/internal/rules/springboot"
)

type dummyVerifier struct{}

func (v *dummyVerifier) Verify(ctx context.Context, repo, dockerfile string, timeout time.Duration) error {
	return nil
}

func TestGenerateAndVerify(t *testing.T) {
	// Set up test environment
	os.Setenv("DG_MOCK_LLM", "1")
	defer os.Unsetenv("DG_MOCK_LLM")

	// Use test directory with pom.xml
	testDir := filepath.Join("testdata")

	client := llm.New()
	verifier := &dummyVerifier{}

	df, err := GenerateAndVerify(context.Background(), testDir, client, verifier)
	if err != nil {
		t.Errorf("GenerateAndVerify() error = %v", err)
		return
	}

	// Check that we got a Dockerfile back
	if df == "" {
		t.Error("GenerateAndVerify() returned empty Dockerfile")
	}
}
