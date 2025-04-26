package main

import (
	"context"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

func main() {
	// Create adapters
	reactDetector := react.NewReactDetector()
	generateAdapter := generate.New()

	// Create Docker driver
	dockerDriver := docker.NewDriver()

	// Create orchestrator
	orchestrator := v2.New(reactDetector, generateAdapter, dockerDriver)

	// Get working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Run pipeline
	if err := orchestrator.Run(context.Background(), dir); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}
