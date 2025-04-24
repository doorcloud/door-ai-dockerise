package dockerverify

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aliou/dockerfile-gen/internal/config"
	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/docker/docker/api/types"
)

type mockDockerClient struct {
	buildCount int
}

func (m *mockDockerClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.buildCount++
	return types.ImageBuildResponse{
		Body: io.NopCloser(strings.NewReader("Successfully built mock-image-id")),
	}, nil
}

// mockLLMClient implements llm.Interface
type mockLLMClient struct{}

var _ llm.Client = (*mockLLMClient)(nil) // Verify interface implementation

func (m *mockLLMClient) AnalyzeFacts(ctx context.Context, snippets []string) (llm.Facts, error) {
	return llm.Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "./mvnw clean package",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Env:       map[string]string{"SPRING_PROFILES_ACTIVE": "production"},
		Health:    "/actuator/health",
	}, nil
}

func (m *mockLLMClient) GenerateDockerfile(ctx context.Context, facts llm.Facts, prevDockerfile string, prevError string) (string, error) {
	return `FROM maven:3.8.4-openjdk-11-slim AS build
WORKDIR /app
COPY . .
RUN chmod +x mvnw && ./mvnw -B -ntp package -DskipTests

FROM openjdk:11-jre-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
}

// copyFile copies a file from src to dst
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

func TestVerifyDockerfile_SkipRebuild(t *testing.T) {
	// Create a mock Docker client
	mockClient := &mockDockerClient{}

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Copy mvnw script from testdata
	mvnwSrc := "testdata/mvnw"
	mvnwDst := filepath.Join(tempDir, "mvnw")
	if err := copyFile(mvnwSrc, mvnwDst); err != nil {
		t.Fatalf("Failed to copy mvnw script: %v", err)
	}
	if err := os.Chmod(mvnwDst, 0755); err != nil {
		t.Fatalf("Failed to make mvnw executable: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		BuildTimeout: 15 * time.Minute,
		Debug:        true,
	}

	// Run the verification with maxAttempts=4
	dockerfile := `FROM maven:3.8.4-openjdk-11-slim AS build
WORKDIR /app
COPY . .
RUN chmod +x mvnw && ./mvnw -B -ntp package -DskipTests

FROM openjdk:11-jre-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`

	ok, logs, err := VerifyWithClient(context.Background(), os.DirFS(tempDir), dockerfile, cfg.BuildTimeout, mockClient)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !ok {
		t.Errorf("Expected verification to succeed, got logs: %s", logs)
	}

	// Verify that we only built once
	if mockClient.buildCount > 1 {
		t.Errorf("Expected at most 1 build, got %d", mockClient.buildCount)
	}
}

func TestVerify(t *testing.T) {
	// Set a shorter timeout for testing
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dockerverify-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy mvnw script from testdata
	mvnwSrc := "testdata/mvnw"
	mvnwDst := filepath.Join(tmpDir, "mvnw")
	if err := copyFile(mvnwSrc, mvnwDst); err != nil {
		t.Fatalf("Failed to copy mvnw script: %v", err)
	}
	if err := os.Chmod(mvnwDst, 0755); err != nil {
		t.Fatalf("Failed to make mvnw executable: %v", err)
	}

	// Create a Dockerfile
	dockerfile := `FROM maven:3.8.4-openjdk-11-slim AS build
WORKDIR /app
COPY . .
RUN ./mvnw -B -ntp package -DskipTests

FROM openjdk:11-jre-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`

	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// Create test config
	cfg := &config.Config{
		BuildTimeout: 15 * time.Minute,
		Debug:        true,
	}

	// Create mock Docker client
	mockClient := &mockDockerClient{}

	// Verify the Dockerfile
	ok, logs, err := VerifyWithClient(ctx, os.DirFS(tmpDir), dockerfile, cfg.BuildTimeout, mockClient)
	if err != nil {
		t.Fatalf("Failed to verify Dockerfile: %v", err)
	}

	// Check the output
	if !ok {
		t.Errorf("Expected verification to succeed, got logs: %s", logs)
	}

	// Verify that we only built once
	if mockClient.buildCount > 1 {
		t.Errorf("Expected at most 1 build, got %d", mockClient.buildCount)
	}
}
