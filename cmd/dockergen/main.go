package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/openai"
)

func main() {
	rootDir := flag.String("root", ".", "Root directory to analyze")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	p := v2.NewPipeline(
		v2.WithLLM(openai.New(os.Getenv("OPENAI_API_KEY"))),
		v2.WithDockerDriver(docker.NewDriver()),
	)
	if err := p.Run(ctx, *rootDir); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}
}
