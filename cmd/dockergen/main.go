package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/loop"
)

func newTestClient() llm.Client {
	return llm.New()
}

func main() {
	// Parse command line arguments
	dir := flag.String("dir", ".", "project directory")
	flag.Parse()

	// Create filesystem for the project directory
	fsys := os.DirFS(*dir)

	// Run the Dockerfile generation loop
	ctx := context.Background()
	client := newTestClient()

	dockerfile, err := loop.Run(ctx, fsys, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write Dockerfile to disk
	if err := os.WriteFile("Dockerfile", []byte(dockerfile), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Dockerfile: %v\n", err)
		os.Exit(1)
	}
}
