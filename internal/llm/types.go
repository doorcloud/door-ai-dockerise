package llm

import (
	"context"
)

// Facts represents the analyzed facts about a project
type Facts struct {
	Language  string
	Framework string
	BuildTool string
	BuildCmd  string
	StartCmd  string
	Ports     []int
	Health    string
	BaseImage string
	Env       map[string]string
	Artifact  string
	BuildDir  string
}

// Client represents an LLM client for analyzing facts and generating Dockerfiles
type Client interface {
	// AnalyzeFacts analyzes code snippets and returns enhanced facts
	AnalyzeFacts(ctx context.Context, snippets []string) (Facts, error)
	// GenerateDockerfile generates a Dockerfile based on the facts
	GenerateDockerfile(ctx context.Context, facts Facts, prevDockerfile string, prevError string) (string, error)
}
