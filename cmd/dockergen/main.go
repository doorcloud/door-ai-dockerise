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
	"github.com/doorcloud/door-ai-dockerise/adapters/docker"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	v1 "github.com/doorcloud/door-ai-dockerise/adapters/orchestrator/v1"
	"github.com/doorcloud/door-ai-dockerise/adapters/spec/loader"
	"github.com/doorcloud/door-ai-dockerise/core"
	coremock "github.com/doorcloud/door-ai-dockerise/core/mock"
	dockerdriver "github.com/doorcloud/door-ai-dockerise/drivers/docker"
)

func main() {
	// Parse flags
	path := flag.String("path", ".", "Path to the project directory")
	specPath := flag.String("spec", "", "Path to stack spec file (yaml/json)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Set debug mode if requested
	if *debug {
		os.Setenv("DG_DEBUG", "1")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Load spec if provided
	var spec *core.Spec
	if *specPath != "" {
		var err error
		spec, err = loader.Load(*specPath)
		if err != nil {
			log.Fatalf("Failed to load spec: %v", err)
		}
	}

	// Create the pipeline components
	generator := generate.NewLLM(coremock.NewMockLLM())
	verifier := docker.NewVerifierAdapter(dockerdriver.NewDriver())

	// Create orchestrator with all components
	o := v1.New(
		[]core.Detector{
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
			node.NewNodeDetector(),
		},
		[]core.FactProvider{
			facts.NewStatic(),
		},
		generator,
		verifier,
	)

	// Run the orchestrator
	dockerfile, err := o.Run(ctx, *path, spec, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	// Write Dockerfile
	if err := os.WriteFile("Dockerfile", []byte(dockerfile), 0644); err != nil {
		log.Fatalf("Failed to write Dockerfile: %v", err)
	}
}
