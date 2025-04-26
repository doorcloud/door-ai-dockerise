package compat

import (
	"context"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// These type aliases let old packages embed the new interfaces gradually.
type Detector = core.Detector
type FactProvider = core.FactProvider
type DockerfileGen = core.DockerfileGen
type Verifier = core.Verifier
type Orchestrator = core.Orchestrator

// Helper no-op implementations for tests (can be deleted later).
type NoopDetector struct{}

func (NoopDetector) Detect(context.Context, string) (core.StackInfo, error) {
	return core.StackInfo{}, nil
}
