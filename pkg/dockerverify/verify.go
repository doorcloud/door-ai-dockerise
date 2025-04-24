package dockerverify

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

// DockerClient defines the interface for Docker operations needed by VerifyDockerfile.
type DockerClient interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

// dockerfileDigest returns the first 12 characters of the SHA256 hash of the Dockerfile content.
func dockerfileDigest(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])[:12]
}

// cleanDockerfile removes markdown formatting from the Dockerfile content.
func cleanDockerfile(content string) string {
	// Remove markdown code block markers and any text before/after the Dockerfile
	lines := strings.Split(content, "\n")
	var cleanLines []string
	inDockerfile := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and markdown code block markers
		if trimmed == "" || trimmed == "```" || trimmed == "```dockerfile" || trimmed == "```Dockerfile" {
			continue
		}

		// Start collecting lines when we see a Dockerfile instruction
		if strings.HasPrefix(trimmed, "FROM ") || strings.HasPrefix(trimmed, "WORKDIR ") ||
			strings.HasPrefix(trimmed, "COPY ") || strings.HasPrefix(trimmed, "RUN ") ||
			strings.HasPrefix(trimmed, "EXPOSE ") || strings.HasPrefix(trimmed, "HEALTHCHECK ") ||
			strings.HasPrefix(trimmed, "ENTRYPOINT ") || strings.HasPrefix(trimmed, "CMD ") {
			inDockerfile = true
		}

		// Stop collecting lines when we see non-Dockerfile content
		if inDockerfile && !strings.HasPrefix(trimmed, "FROM ") && !strings.HasPrefix(trimmed, "WORKDIR ") &&
			!strings.HasPrefix(trimmed, "COPY ") && !strings.HasPrefix(trimmed, "RUN ") &&
			!strings.HasPrefix(trimmed, "EXPOSE ") && !strings.HasPrefix(trimmed, "HEALTHCHECK ") &&
			!strings.HasPrefix(trimmed, "ENTRYPOINT ") && !strings.HasPrefix(trimmed, "CMD ") &&
			!strings.HasPrefix(trimmed, "#") {
			break
		}

		if inDockerfile {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// createDockerignore creates a .dockerignore file in the build context directory
func createDockerignore(dir string) error {
	ignoreContent := `.git
**/*.iml
.idea
docs
*.md
!mvnw
!.mvn/**
`
	return os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(ignoreContent), 0644)
}

// VerifyDockerfile attempts to build and run a Dockerfile, with retries on failure.
func VerifyDockerfile(ctx context.Context, cli DockerClient, repo string, f facts.Facts, llmClient llm.Interface, maxAttempts int, cfg *config.Config) (string, error) {
	var dockerfile string
	var err error
	var lastDigest string
	var imageID string

	log.Printf("Starting Dockerfile verification process for repository: %s", repo)
	log.Printf("Maximum attempts: %d", maxAttempts)

	// Use build timeout from config
	buildTimeout := cfg.BuildTimeout
	log.Printf("Build timeout set to: %v", buildTimeout)

	// Convert facts to map for LLM
	factsMap := f.ToMap()
	log.Printf("Facts map created: %+v", factsMap)

	// Get Maven cache directory from config
	m2Cache := cfg.GetM2Cache()
	if _, err := os.Stat(m2Cache); err != nil {
		log.Printf("Warning: Maven cache directory not found at %s", m2Cache)
		m2Cache = ""
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempt %d/%d", attempt, maxAttempts)

		// Create a context with timeout for LLM operations
		llmCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		if attempt > 1 {
			log.Printf("Retrying with error feedback from previous attempt")
			// Get error log from previous attempt
			buildErrLog := collectLast100Lines(err)
			log.Printf("Previous error log (last 100 lines):\n%s", buildErrLog)

			// Determine error type
			errorType := "unknown"
			if strings.Contains(buildErrLog, "pom.xml") {
				errorType = "missing_pom"
			} else if strings.Contains(buildErrLog, "JDK") || strings.Contains(buildErrLog, "Java") {
				errorType = "jdk_version"
			} else if strings.Contains(buildErrLog, "test") {
				errorType = "test_failure"
			} else if strings.Contains(buildErrLog, "dependency") {
				errorType = "dependency_error"
			}

			// Log the current Dockerfile that failed
			log.Printf("Current Dockerfile that failed:\n%s", dockerfile)

			// Log the LLM retry request
			log.Printf("Sending retry request to LLM with:")
			log.Printf("- Facts: %+v", factsMap)
			log.Printf("- Current Dockerfile: %s", dockerfile)
			log.Printf("- Repository: %s", repo)
			log.Printf("- Error log: %s", buildErrLog)
			log.Printf("- Error type: %s", errorType)
			log.Printf("- Attempt number: %d", attempt)

			dockerfile, err = llmClient.FixDockerfile(llmCtx, factsMap, dockerfile, repo, buildErrLog, errorType, attempt)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("Context cancelled during LLM retry: %v", ctx.Err())
					return "", fmt.Errorf("context cancelled during LLM retry: %w", ctx.Err())
				}
				log.Printf("LLM retry failed: %v", err)
				return "", fmt.Errorf("LLM retry %d: %w", attempt, err)
			}
			log.Printf("Dockerfile fixed successfully in attempt %d", attempt)
			log.Printf("New Dockerfile:\n%s", dockerfile)
		} else {
			log.Printf("Generating initial Dockerfile")
			dockerfile, err = llmClient.GenerateDockerfile(llmCtx, factsMap)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("Context cancelled during Dockerfile generation: %v", ctx.Err())
					return "", fmt.Errorf("context cancelled during Dockerfile generation: %w", ctx.Err())
				}
				log.Printf("Dockerfile generation failed: %v", err)
				return "", fmt.Errorf("generate Dockerfile: %w", err)
			}
			log.Printf("Initial Dockerfile generated successfully")
			log.Printf("Generated Dockerfile:\n%s", dockerfile)
		}

		// Clean the Dockerfile content
		dockerfile = cleanDockerfile(dockerfile)

		// Create temporary directory for build
		tmpDir, err := os.MkdirTemp("", "dockergen-*")
		if err != nil {
			log.Printf("Failed to create temp dir: %v", err)
			return "", fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)
		log.Printf("Created temporary build directory: %s", tmpDir)

		// Write Dockerfile
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
			log.Printf("Failed to write Dockerfile: %v", err)
			return "", fmt.Errorf("write Dockerfile: %w", err)
		}
		log.Printf("Dockerfile written to: %s", dockerfilePath)

		// Create .dockerignore
		if err := createDockerignore(tmpDir); err != nil {
			log.Printf("Failed to create .dockerignore: %v", err)
			return "", fmt.Errorf("create .dockerignore: %w", err)
		}

		// Copy build context
		log.Printf("Copying build context from %s to %s", repo, tmpDir)
		if err := copyDir(repo, tmpDir); err != nil {
			log.Printf("Failed to copy build context: %v", err)
			return "", fmt.Errorf("copy build context: %w", err)
		}
		log.Printf("Build context copied successfully")

		// Fix mvnw permissions if it exists
		mvnwPath := filepath.Join(tmpDir, "mvnw")
		if _, err := os.Stat(mvnwPath); err == nil {
			log.Printf("Found mvnw script, fixing permissions")
			if err := os.Chmod(mvnwPath, 0755); err != nil {
				log.Printf("Failed to fix mvnw permissions: %v", err)
				return "", fmt.Errorf("fix mvnw permissions: %w", err)
			}
			log.Printf("Fixed mvnw permissions to 0755")
		}

		// Compute Dockerfile digest
		currentDigest := dockerfileDigest([]byte(dockerfile))
		imageTag := fmt.Sprintf("dockergen-e2e:%s", currentDigest)

		// Only build if this is the first attempt or if the Dockerfile has changed
		if attempt == 1 || currentDigest != lastDigest {
			// Build with timeout
			buildCtx, cancel := context.WithTimeout(ctx, buildTimeout)
			defer cancel()
			log.Printf("Starting Docker build with %v timeout", buildTimeout)

			// Prepare build args for Maven cache
			buildArgs := []string{"build", "-t", imageTag}
			if m2Cache != "" {
				buildArgs = append(buildArgs, "--build-arg", "BUILDKIT_INLINE_CACHE=1")
				buildArgs = append(buildArgs, "--secret", "id=m2,target=/root/.m2")
			}
			buildArgs = append(buildArgs, ".")

			cmd := exec.CommandContext(buildCtx, "docker", buildArgs...)
			cmd.Dir = tmpDir
			cmd.Env = append(os.Environ(),
				"DOCKER_BUILDKIT=1",
				"BUILDKIT_PROGRESS=plain",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("Context cancelled during build: %v", ctx.Err())
					return "", fmt.Errorf("context cancelled during build: %w", ctx.Err())
				}
				log.Printf("Build failed: %v\nOutput:\n%s", err, string(output))
				err = fmt.Errorf("build failed: %w\n%s", err, string(output))
				continue
			}

			// Extract image ID from BuildKit output
			lines := strings.Split(string(output), "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				if strings.HasPrefix(lines[i], "sha256:") {
					imageID = strings.TrimSpace(lines[i])
					break
				}
			}
			log.Printf("Docker build completed successfully, image ID: %s", imageID)
		} else {
			log.Printf("Skipping build, using existing image: %s", imageID)
		}

		// Run with timeout
		runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		log.Printf("Starting Docker run with 2-minute timeout")

		// Run container in detached mode
		containerName := fmt.Sprintf("dockergen-verify-%s", currentDigest)
		cmd := exec.CommandContext(runCtx, "docker", "run", "-d", "--rm", "--name", containerName, imageTag)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("Context cancelled during run: %v", ctx.Err())
				return "", fmt.Errorf("context cancelled during run: %w", ctx.Err())
			}
			log.Printf("Run failed: %v\nOutput:\n%s", err, string(output))
			err = fmt.Errorf("run failed: %w\n%s", err, string(output))
			continue
		}

		// Wait for container to start
		time.Sleep(5 * time.Second)

		// Get container IP
		cmd = exec.CommandContext(runCtx, "docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Failed to get container IP: %v\nOutput:\n%s", err, string(output))
			continue
		}
		containerIP := strings.TrimSpace(string(output))

		// Check health endpoint if provided
		if f.Health != "" {
			healthURL := fmt.Sprintf("http://%s:8080%s", containerIP, f.Health)
			log.Printf("Checking health endpoint: %s", healthURL)

			// Create HTTP client with timeout
			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			// Try health check up to 3 times
			var lastErr error
			for i := 0; i < 3; i++ {
				resp, err := client.Get(healthURL)
				if err != nil {
					lastErr = err
					time.Sleep(5 * time.Second)
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					lastErr = fmt.Errorf("health check failed with status: %d", resp.StatusCode)
					time.Sleep(5 * time.Second)
					continue
				}

				// Health check passed
				lastErr = nil
				break
			}

			if lastErr != nil {
				log.Printf("Health check failed: %v", lastErr)
				err = fmt.Errorf("health check failed: %w", lastErr)
				continue
			}
			log.Printf("Health check passed")
		}

		// Stop container
		cmd = exec.CommandContext(runCtx, "docker", "stop", containerName)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to stop container: %v", err)
		}

		lastDigest = currentDigest
		return dockerfile, nil
	}

	log.Printf("Verification failed after %d attempts", maxAttempts)
	return "", fmt.Errorf("verify failed after %d attempts", maxAttempts)
}

// collectLast100Lines returns the last 100 lines of the error output.
func collectLast100Lines(err error) string {
	if err == nil {
		return ""
	}
	lines := strings.Split(err.Error(), "\n")
	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}
	return strings.Join(lines, "\n")
}

// copyDir copies a directory recursively.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
