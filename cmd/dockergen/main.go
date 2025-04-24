package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aliou/dockerfile-gen/internal/config"
	"github.com/aliou/dockerfile-gen/internal/generator"
	"github.com/aliou/dockerfile-gen/internal/llm"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repository-path>\n", os.Args[0])
		os.Exit(1)
	}

	// Get repository path
	repoPath := os.Args[1]

	// Load configuration
	cfg := config.New()

	// Create LLM client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY environment variable is required")
		os.Exit(1)
	}

	client, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create LLM client: %v\n", err)
		os.Exit(1)
	}

	// Open repository as fs.FS
	fsys := os.DirFS(repoPath)

	// Generate Dockerfile
	ctx := context.Background()
	dockerfile, err := generator.Generate(ctx, fsys, client, 3, cfg.BuildTimeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate Dockerfile: %v\n", err)
		os.Exit(1)
	}

	// Write Dockerfile to stdout
	fmt.Print(dockerfile)
}
