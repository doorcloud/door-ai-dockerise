package main

import (
	"fmt"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
)

func main() {
	// Enable verbose logging if DEBUG=true
	if os.Getenv("DEBUG") == "true" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repo-path>\n", os.Args[0])
		os.Exit(1)
	}

	// Initialize LLM client
	client := llm.New()

	// Run the pipeline
	if err := pipeline.Run(os.Args[1], client); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
