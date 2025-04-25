package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/doorcloud/door-ai-dockerise/internal/verify"
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

	// Get filesystem
	fsys := os.DirFS(os.Args[1])

	// Detect project type using registry
	reg := rules.NewRegistry()
	rule, ok := reg.Detect(fsys)
	if !ok {
		fmt.Fprintln(os.Stderr, "No matching rule found")
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
