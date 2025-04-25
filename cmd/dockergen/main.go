package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/springboot"
)

func main() {
	// Parse command line arguments
	repo := flag.String("repo", ".", "path to the repository")
	flag.Parse()

	// Create a new registry
	registry := rules.NewRegistry()

	// Register the Spring Boot detector
	registry.Register(&springboot.SpringBoot{})

	// Create a filesystem for the repository
	fsys := os.DirFS(*repo)

	// Detect the technology stack
	rule, detected := registry.Detect(fsys)
	if !detected {
		fmt.Println("No technology stack detected")
		os.Exit(1)
	}

	// Print the detected stack
	fmt.Printf("Detected stack: %s\n", rule.Name)
}
