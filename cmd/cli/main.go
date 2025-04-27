package main

import (
	"context"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/pipeline"
)

func main() {
	// Create context
	ctx := context.Background()

	// Create mock LLM for testing
	mockLLM := mock.NewMockLLM()

	// Create pipeline with all components
	p := pipeline.NewPipeline(
		pipeline.WithDetectors(
			react.NewReactDetector(),
			springboot.NewSpringBootDetector(),
		),
		pipeline.WithFactProviders(
			facts.NewStatic(),
		),
		pipeline.WithGenerator(generate.NewLLM(mockLLM)),
		pipeline.WithDockerDriver(docker.NewDriver()),
		pipeline.WithMaxRetries(3),
	)

	// Get source path from command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Please provide a source path")
	}
	sourcePath := os.Args[1]

	// Run the pipeline
	if err := p.Run(ctx, sourcePath); err != nil {
		log.Fatal(err)
	}
}
