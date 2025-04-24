package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/pkg/dockerverify"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments
	repoPath := flag.String("repo", "", "Path to repository")
	flag.Parse()

	if *repoPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -repo is required")
		os.Exit(1)
	}

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create LLM client
	llmClient, err := llm.NewClient(&cfg)
	if err != nil {
		logger.Error("Failed to create LLM client", "error", err)
		os.Exit(1)
	}

	// Load rule configs
	configs, err := rules.LoadRuleConfigs("internal/rules/configs")
	if err != nil {
		logger.Error("Failed to load rule configs", "error", err)
		os.Exit(1)
	}

	config, ok := configs["springboot"]
	if !ok {
		logger.Error("SpringBoot config not found")
		os.Exit(1)
	}

	// Create SpringBoot rule
	rule := springboot.NewRule(logger, llmClient, config)

	// Extract snippets
	snippets, err := rule.Snippets(*repoPath)
	if err != nil {
		logger.Error("Failed to extract snippets", "error", err)
		os.Exit(1)
	}

	// Extract facts
	ctx := context.Background()
	projectFacts, err := rule.Facts(ctx, snippets, llmClient)
	if err != nil {
		logger.Error("Failed to extract facts", "error", err)
		os.Exit(1)
	}

	// Generate initial Dockerfile
	dockerfile, err := llmClient.GenerateDockerfile(ctx, projectFacts.ToMap())
	if err != nil {
		logger.Error("Failed to generate Dockerfile", "error", err)
		os.Exit(1)
	}

	// Write initial Dockerfile
	dockerfilePath := filepath.Join(*repoPath, "Dockerfile.generated")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		logger.Error("Failed to write Dockerfile", "error", err)
		os.Exit(1)
	}

	// Create Docker client for verification
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logger.Error("Failed to create Docker client", "error", err)
		os.Exit(1)
	}

	// Verify and retry loop
	var lastErrorType string
	for attempt := 1; attempt <= 4; attempt++ {
		if cfg.Debug {
			logger.Info("Verifying Dockerfile", "attempt", attempt)
		}

		_, err := dockerverify.VerifyDockerfile(ctx, dockerClient, *repoPath, projectFacts, llmClient, attempt, &cfg)
		if err == nil {
			logger.Info("Dockerfile verification successful", "attempt", attempt)
			break
		}

		// Get error synopsis
		errLines := strings.Split(err.Error(), "\n")
		synopsis := strings.Join(errLines[:min(3, len(errLines))], "\n")

		// Check for pom.xml locations
		findCmd := exec.Command("find", ".", "-name", "pom.xml")
		findCmd.Dir = *repoPath
		pomOutput, _ := findCmd.Output()

		logger.Info("Dockerfile verification failed",
			"attempt", attempt,
			"error", synopsis,
			"pom_files", string(pomOutput))

		// Check if error type changed
		errorType := getErrorType(err.Error())
		if lastErrorType != "" && errorType != lastErrorType {
			logger.Info("Error type changed, stopping retry loop",
				"previous", lastErrorType,
				"current", errorType)
			break
		}
		lastErrorType = errorType

		if attempt == 4 {
			logger.Error("Dockerfile verification failed after 4 attempts", "error", err)
			os.Exit(1)
		}

		if cfg.Debug {
			logger.Info("Attempting to fix Dockerfile", "attempt", attempt)
		}

		// Get error log from verify
		errLog := err.Error()

		// Generate fixed Dockerfile
		dockerfile, err = llmClient.FixDockerfile(ctx, projectFacts.ToMap(), dockerfile, *repoPath, errLog, errorType, attempt)
		if err != nil {
			logger.Error("Failed to fix Dockerfile", "error", err)
			os.Exit(1)
		}

		// Write updated Dockerfile
		if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
			logger.Error("Failed to write Dockerfile", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("Successfully generated and verified Dockerfile", "path", dockerfilePath)
}

// Helper function to get error type
func getErrorType(errStr string) string {
	if strings.Contains(errStr, "no POM in this directory") {
		return "missing_pom"
	}
	if strings.Contains(errStr, "Java version") {
		return "java_version"
	}
	if strings.Contains(errStr, "mvnw") {
		return "mvnw_missing"
	}
	return "other"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
