package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/internal/build"
	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/prompt"
)

func main() {
	repo := flag.String("repo", ".", "path to repository")
	tag := flag.String("tag", "spring-app:latest", "docker image tag")
	retry := flag.Int("retry", 0, "number of retries on build failure")
	flag.Parse()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY env var is required")
	}

	absRepo, _ := filepath.Abs(*repo)
	if detect.Detect(absRepo) != "spring" {
		log.Fatal("unsupported repo: Spring Boot not detected")
	}

	p := prompt.Render(filepath.Base(absRepo))
	dockerfile, err := llm.Generate(p, apiKey)
	if err != nil {
		log.Fatalf("LLM error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(absRepo, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		log.Fatalf("cannot write Dockerfile: %v", err)
	}

	for attempt := 0; ; attempt++ {
		if err := build.Build(absRepo, *tag); err == nil {
			fmt.Println("🟢 Image built:", *tag)
			return
		}
		if attempt >= *retry {
			log.Fatalf("build failed after %d attempts", attempt+1)
		}
		fmt.Printf("⚠️  Build failed (attempt %d/%d), retrying...\n", attempt+1, *retry)
	}
}
