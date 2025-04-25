package pipeline

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/pkg/dockerverify"
)

// Attempts is hard-coded to 4 for now.
const attempts = 4

// GenerateAndVerify runs the full flow and returns the *final* Dockerfile.
func GenerateAndVerify(ctx context.Context, repo string, c llm.Client, v dockerverify.Verifier) (string, error) {
	// 1) Facts ─────────────────────────────────────────────
	if _, err := c.Chat("facts", repo); err != nil {
		return "", err
	}

	// 2) Rule ──────────────────────────────────────────────
	if _, err := rules.DetectStack(os.DirFS(repo)); err != nil {
		return "", err
	}

	// 3) Dockerfile loop ──────────────────────────────────
	df := ""
	for i := 0; i < attempts; i++ {
		var err error
		if i == 0 {
			df, err = c.Chat("dockerfile", repo)
		} else {
			df, err = c.Chat("dockerfile", repo+"\nError: "+err.Error()+"\nCurrent Dockerfile:\n"+df)
		}
		if err != nil {
			return "", err
		}

		if err = v.Verify(ctx, repo, df, 20*time.Minute); err == nil {
			return df, nil // success 🎉
		}
	}

	return "", errors.New("verify failed after 4 attempts")
}
