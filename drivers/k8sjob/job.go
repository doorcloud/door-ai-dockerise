//go:build k8s
// +build k8s

package k8sjob

import (
	"context"
	"fmt"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// DummyJob implements core.BuildDriver for k8s jobs
type DummyJob struct{}

// NewJob creates a new k8s job driver
func NewJob() core.BuildDriver {
	return &DummyJob{}
}

// Build implements core.BuildDriver
func (j *DummyJob) Build(ctx context.Context, in core.BuildInput, log core.LogStreamer) (core.ImageRef, error) {
	// TODO: Implement k8s job build
	log.Info("Starting k8s job build")
	log.Error("k8s job build not implemented")
	return core.ImageRef{}, fmt.Errorf("k8s job build not implemented")
}
