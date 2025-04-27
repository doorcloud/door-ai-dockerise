package main

import (
	"context"
	"flag"
	"log"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/pipeline"
	"github.com/doorcloud/door-ai-dockerise/providers/facts"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

func main() {
	dir := flag.String("dir", ".", "Directory to analyze")
	apiKey := flag.String("api-key", "", "OpenAI API key")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("OpenAI API key is required")
	}

	llmClient := openai.NewProvider(*apiKey)
	dockerDriver := docker.NewMockDriver()

	// Create pipeline with detectors
	p := pipeline.NewPipeline(
		pipeline.WithDetectors(detectors.DefaultDetectors()...),
		pipeline.WithFactProviders(facts.DefaultProviders(llmClient)...),
		pipeline.WithGenerator(llmClient),
		pipeline.WithDockerDriver(dockerDriver),
		pipeline.WithMaxRetries(3),
	)

	// Run the pipeline
	if err := p.Run(context.Background(), *dir); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}
