package dockerverify

import (
	"context"
	"time"
)

// Verifier defines the interface for verifying Dockerfiles
type Verifier interface {
	Verify(ctx context.Context, repo, dockerfile string, timeout time.Duration) error
}
