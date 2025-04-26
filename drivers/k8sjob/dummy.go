//go:build !k8s
// +build !k8s

package k8sjob

import (
	"context"
	"io"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// DummyJob is a placeholder for non-k8s builds
type DummyJob struct{}

// NewJob returns a dummy job for non-k8s builds
func NewJob() (*DummyJob, error) {
	return &DummyJob{}, nil
}

// Build implements the core.BuildDriver interface
func (j *DummyJob) Build(ctx context.Context, in core.BuildInput, w io.Writer) (core.ImageRef, error) {
	return core.ImageRef{}, nil
}
