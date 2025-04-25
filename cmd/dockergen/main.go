package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aliou/dockerfile-gen/internal/detect"
	"github.com/aliou/dockerfile-gen/internal/facts"
	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/aliou/dockerfile-gen/internal/types"
	"github.com/aliou/dockerfile-gen/internal/verify"
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

	// Get filesystem
	fsys := os.DirFS(os.Args[1])

	// Detect project type
	rule, err := detect.Detect(fsys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting project type: %v\n", err)
		os.Exit(1)
	}

	// Infer facts about the project
	projectFacts, err := facts.InferWithClient(context.Background(), fsys, rule, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inferring facts: %v\n", err)
		os.Exit(1)
	}

	// Convert facts to types.Facts
	typedFacts := types.Facts{
		Language:  projectFacts.Language,
		Framework: projectFacts.Framework,
		BuildTool: projectFacts.BuildTool,
		BuildCmd:  projectFacts.BuildCmd,
		BuildDir:  projectFacts.BuildDir,
		StartCmd:  projectFacts.StartCmd,
		Artifact:  projectFacts.Artifact,
		Ports:     projectFacts.Ports,
		Health:    projectFacts.Health,
		Env:       projectFacts.Env,
		BaseImage: projectFacts.BaseImage,
	}

	// Generate and verify Dockerfile
	var dockerfile string
	var errLog string
	for i := 0; i < 3; i++ {
		// Generate Dockerfile
		dockerfile, err = llm.GenerateDockerfile(context.Background(), typedFacts, dockerfile, errLog, i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating Dockerfile: %v\n", err)
			os.Exit(1)
		}

		// Verify the Dockerfile
		if err := verify.Verify(context.Background(), fsys, dockerfile); err == nil {
			break // Success!
		} else {
			errLog = err.Error()
			if i == 2 {
				fmt.Fprintf(os.Stderr, "Failed to generate valid Dockerfile after 3 attempts: %v\n", err)
				os.Exit(1)
			}
		}
	}

	fmt.Println(dockerfile)
}
