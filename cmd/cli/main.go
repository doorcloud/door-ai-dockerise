package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
)

func main() {
	// Parse command line flags
	projectDir := flag.String("dir", ".", "Project directory to analyze")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create mock LLM for testing
	mockLLM := mock.NewMockLLM()

	// Create pipeline
	p := pipeline.New(cfg,
		pipeline.WithDetectors(
			detectors.NewReactDetector(),
			detectors.NewSpringBootDetector(),
		),
		pipeline.WithFactProviders(
			facts.NewStatic(),
		),
		pipeline.WithGenerator(generate.NewLLM(mockLLM)),
		pipeline.WithDockerDriver(docker.New()),
	)

	// Run pipeline
	if err := p.Run(context.Background(), *projectDir, os.Stdout); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}
