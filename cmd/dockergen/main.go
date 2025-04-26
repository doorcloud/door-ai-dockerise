package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/node"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers/docker"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/pipeline/v2"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

func main() {
	rootDir := flag.String("root", ".", "Root directory to analyze")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	p := defaultPipeline()
	if err := p.Run(ctx, *rootDir); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}

func defaultPipeline() *pipeline.Orchestrator {
	return pipeline.New(
		core.DetectorChain{
			react.New(),
			springboot.New(),
			node.New(),
		},
		openai.New(os.Getenv("OPENAI_API_KEY")),
		docker.New(),
	)
}
