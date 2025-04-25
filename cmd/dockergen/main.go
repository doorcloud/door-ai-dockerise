package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aliou/dockerfile-gen/internal/generate"
	"github.com/aliou/dockerfile-gen/internal/llm"
)

func main() {
	// Parse command line flags
	repo := flag.String("repo", "", "Path to the repository")
	maxRetries := flag.Int("max-retries", 3, "Maximum number of retries for Dockerfile generation")
	buildTimeout := flag.Duration("build-timeout", 5*time.Minute, "Timeout for Docker build verification")
	flag.Parse()

	if *repo == "" {
		log.Fatal("repository path is required")
	}

	// Create filesystem from repository path
	fsys := os.DirFS(*repo)

	// Initialize LLM client
	cli, err := llm.NewClient()
	if err != nil {
		log.Fatalf("failed to create LLM client: %v", err)
	}

	// Generate Dockerfile
	dockerfile, err := generate.Generate(context.Background(), fsys, cli, *maxRetries, *buildTimeout)
	if err != nil {
		log.Fatalf("failed to generate Dockerfile: %v", err)
	}

	// Print the generated Dockerfile
	fmt.Println(dockerfile)
}
