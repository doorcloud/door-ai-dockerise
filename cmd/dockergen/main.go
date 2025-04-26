package main

import (
	"context"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

func main() {
	// Create context
	ctx := context.Background()

	// Create mock LLM for testing
	mockLLM := mock.NewMockLLM()

	// Create pipeline with all components
	p := v2.New(v2.Options{
		Detectors: []core.Detector{
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		},
		FactProviders: []core.FactProvider{
			facts.NewStatic(),
		},
		Generator:  generate.NewLLM(mockLLM),
		Verifier:   docker.NewDriver(),
		MaxRetries: 3,
	})

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
