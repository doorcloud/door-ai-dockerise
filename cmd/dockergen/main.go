package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	_ "github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
)

func main() {
	// Parse command line flags
	confidenceMin := flag.Float64("detect-confidence-min", 0.5, "Minimum confidence threshold for stack detection")
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
		pipeline.WithDetectors(detectors.Registry()...),
		pipeline.WithFactProviders(
			facts.NewStatic(),
		),
		pipeline.WithGenerator(generate.NewLLM(mockLLM)),
		pipeline.WithDockerDriver(docker.New()),
	)

	// Set confidence threshold
	p.SetConfidenceThreshold(*confidenceMin)

	// Get current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Run pipeline
	ctx := context.Background()
	if err := p.Run(ctx, dir, os.Stdout); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}
