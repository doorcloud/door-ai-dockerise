package pipeline

import (
	"context"
	"errors"
	"time"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/registry"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/pkg/dockerverify"
)

// Attempts is hard-coded to 4 for now.
const attempts = 4

// GenerateAndVerify runs the full flow and returns the *final* Dockerfile.
func GenerateAndVerify(ctx context.Context, repo string, c llm.Client, v dockerverify.Verifier) (string, error) {
	// 1) Facts ─────────────────────────────────────────────
	facts, err := c.Chat("facts", repo)
	if err != nil {
		return "", err
	}

	// 2) Rule ──────────────────────────────────────────────
	var chosen rules.Rule
	for _, r := range registry.All() {
		if r.Match(facts) {
			chosen = r
			break
		}
	}
	if chosen == nil {
		return "", errors.New("no rule matched")
	}

	// 3) Dockerfile loop ──────────────────────────────────
	df := ""
	for i := 0; i < attempts; i++ {
		if i == 0 {
			df, err = c.Chat("dockerfile", chosen.GenPrompt())
		} else {
			df, err = c.Chat("dockerfile", chosen.FixPrompt())
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
