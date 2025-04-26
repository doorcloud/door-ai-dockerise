package v2

import (
	"context"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

// Orchestrator coordinates the Dockerfile generation pipeline
type Orchestrator struct {
	detector       core.Detector
	chatCompletion core.ChatCompletion
	dockerDriver   docker.Driver
}

// New creates a new orchestrator
func New(detector core.Detector, chatCompletion core.ChatCompletion, dockerDriver docker.Driver) *Orchestrator {
	return &Orchestrator{
		detector:       detector,
		chatCompletion: chatCompletion,
		dockerDriver:   dockerDriver,
	}
}

// Run executes the Dockerfile generation pipeline
func (o *Orchestrator) Run(ctx context.Context, dir string) error {
	// Convert directory to filesystem
	fsys := os.DirFS(dir)

	// Detect stack type
	stack, err := o.detector.Detect(ctx, fsys)
	if err != nil {
		return err
	}

	// Gather facts about the stack
	facts, err := o.chatCompletion.GatherFacts(ctx, fsys, stack)
	if err != nil {
		return err
	}

	// Generate Dockerfile
	dockerfile, err := o.chatCompletion.GenerateDockerfile(ctx, facts)
	if err != nil {
		return err
	}

	// Write Dockerfile
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return err
	}

	// Build Docker image
	opts := docker.BuildOptions{
		Context:    dir,
		Tags:       []string{"myapp:latest"},
		Dockerfile: "Dockerfile",
	}
	if err := o.dockerDriver.Build(ctx, dockerfilePath, opts); err != nil {
		return err
	}

	return nil
}
