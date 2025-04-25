package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/aliou/dockerfile-gen/internal/loop"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <repo-path>\n", os.Args[0])
		os.Exit(1)
	}

	// Initialize LLM client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY is required")
		os.Exit(1)
	}
	client := llm.NewClient(apiKey)

	// Run the generation loop
	dockerfile, err := loop.Run(context.Background(), os.DirFS(os.Args[1]), client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(dockerfile)
}
