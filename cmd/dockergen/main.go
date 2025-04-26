package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/node"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
)

func main() {
	// Parse flags
	path := flag.String("path", ".", "Path to the project directory")
	debug := flag.Bool("debug", false, "Enable debug logging")
	attempts := flag.Int("attempts", 3, "Maximum number of generation attempts")
	flag.Parse()

	// Set debug mode if requested
	if *debug {
		os.Setenv("DG_DEBUG", "1")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create pipeline with all components
	p := v2.New(v2.Options{
		Detectors: []core.Detector{
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
			node.NewNodeDetector(),
		},
		FactProviders: []core.FactProvider{
			facts.NewStatic(),
		},
		Generator:  generate.NewLLM(mock.NewMockLLM()),
		Verifier:   docker.NewDriver(),
		MaxRetries: *attempts,
	})

	// Run the pipeline
	if err := p.Run(ctx, *path); err != nil {
		log.Fatal(err)
	}
}
