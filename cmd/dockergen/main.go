package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/pipeline"
	"github.com/doorcloud/door-ai-dockerise/pkg/dockerverify"
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

	// Generate and verify Dockerfile
	df, err := pipeline.GenerateAndVerify(context.Background(), os.Args[1], client, dockerverify.New())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(df)
}
