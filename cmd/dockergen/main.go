package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers"
	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
	"github.com/doorcloud/door-ai-dockerise/providers/facts"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

func main() {
	dir := flag.String("dir", ".", "Directory to analyze")
	apiKey := flag.String("api-key", "", "OpenAI API key")
	flag.Parse()

	if *apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: API key is required")
		os.Exit(1)
	}

	llmClient := openai.NewProvider(*apiKey)
	verifier, err := verifiers.NewDocker(verifiers.Options{
		Socket:  "unix:///var/run/docker.sock",
		LogSink: os.Stdout,
		Timeout: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to create Docker verifier: %v", err)
	}

	p := pipeline.New(pipeline.Options{
		Detectors:     detectors.DefaultDetectors(),
		FactProviders: facts.DefaultProviders(llmClient),
		Generator:     llmClient,
		Verifier:      verifier,
		MaxAttempts:   3,
	})

	dockerfile, err := p.Process(context.Background(), *dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(dockerfile)
}
