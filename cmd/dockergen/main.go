package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
)

func main() {
	// Parse flags
	dir := flag.String("dir", ".", "Directory containing the project")
	flag.Parse()

	// Create pipeline
	p := pipeline.New(*dir)

	// Run pipeline
	dockerfile, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Write Dockerfile
	if err := os.WriteFile("Dockerfile", []byte(dockerfile), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Dockerfile: %v\n", err)
		os.Exit(1)
	}
}
