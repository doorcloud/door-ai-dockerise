//go:build !k8s
// +build !k8s

package k8sjob

import (
	"context"
	"fmt"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// DummyJob implements core.BuildDriver for non-k8s builds
type DummyJob struct{}

// NewJob returns a dummy job for non-k8s builds
func NewJob() core.BuildDriver {
	return &DummyJob{}
}

// Build implements core.BuildDriver
func (j *DummyJob) Build(ctx context.Context, in core.BuildInput, log core.LogStreamer) (core.ImageRef, error) {
	log.Error("k8s job build not available - build with k8s tag")
	return core.ImageRef{}, fmt.Errorf("k8s job build not available - build with k8s tag")
}
