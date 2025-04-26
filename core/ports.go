package core

import "context"

// StackInfo is a minimal description of the detected tech stack.
type StackInfo struct {
	Name    string            // e.g. "react", "spring-boot"
	Version string            // optional semantic version
	Meta    map[string]string // extra info (build tool, language…)
}

// Detector analyses a path and returns the stack details.
type Detector interface {
	Detect(ctx context.Context, path string) (StackInfo, error)
}

// FactProvider enriches a stack with structured facts.
type FactProvider interface {
	Facts(ctx context.Context, stack StackInfo) ([]string, error)
}

// DockerfileGen turns facts into a Dockerfile string.
type DockerfileGen interface {
	Generate(ctx context.Context, facts []string) (string, error)
}

// Verifier builds / runs the image and returns an error if it fails.
type Verifier interface {
	Verify(ctx context.Context, repoPath, dockerfile string) error
}

// Orchestrator ties the above pieces together.
type Orchestrator interface {
	Run(ctx context.Context, repoPath string) (string /*final Dockerfile*/, error)
}
