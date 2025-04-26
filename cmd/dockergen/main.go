package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/adapters/log/json"
	"github.com/doorcloud/door-ai-dockerise/adapters/log/plain"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/orchestrator"
)

func main() {
	// Parse flags
	root := flag.String("root", ".", "Project root directory")
	logFormat := flag.String("log", "plain", "Log format (plain|json)")
	flag.Parse()

	// Create log streamer
	var log core.LogStreamer
	switch *logFormat {
	case "plain":
		log = plain.New()
	case "json":
		log = json.New()
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format: %s\n", *logFormat)
		os.Exit(1)
	}

	// Create orchestrator
	o := orchestrator.New(orchestrator.Opts{
		Log: nil, // We'll use the LogStreamer directly in Run
	})

	// Run the orchestrator
	ctx := context.Background()
	dockerfile, err := o.Run(ctx, *root, nil, log)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to generate Dockerfile: %v", err))
		os.Exit(1)
	}

	// Print the generated Dockerfile
	fmt.Println(dockerfile)
}
