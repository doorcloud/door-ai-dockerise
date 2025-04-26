package verifiers

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/verify"
)

type Docker struct{}

func NewDocker() *Docker {
	return &Docker{}
}

func (d *Docker) Verify(ctx context.Context, repoPath, dockerfile string) error {
	fsys := os.DirFS(repoPath)
	return verify.Verify(ctx, fsys, dockerfile)
}
