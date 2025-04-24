package dockerverify

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aliou/dockerfile-gen/internal/config"
	"github.com/aliou/dockerfile-gen/internal/llm"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockDockerClient struct {
	buildCount int
	containers map[string]bool
}

func (m *mockDockerClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	return []types.ImageSummary{}, nil
}

func (m *mockDockerClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.buildCount++
	return types.ImageBuildResponse{
		Body: io.NopCloser(strings.NewReader("Successfully built mock-image-id")),
	}, nil
}

func (m *mockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	if m.containers == nil {
		m.containers = make(map[string]bool)
	}
	m.containers[containerName] = true
	return container.CreateResponse{
		ID: containerName,
	}, nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	if !m.containers[containerID] {
		return fmt.Errorf("No such object: %s", containerID)
	}
	return nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	if !m.containers[containerID] {
		return fmt.Errorf("No such object: %s", containerID)
	}
	delete(m.containers, containerID)
	return nil
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if !m.containers[containerID] {
		return types.ContainerJSON{}, fmt.Errorf("No such object: %s", containerID)
	}
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Status: "running",
			},
		},
		NetworkSettings: &types.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"bridge": {
					IPAddress: "172.17.0.2",
				},
			},
		},
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

func TestVerifyDockerfile_SkipRebuild(t *testing.T) {
	// Create a mock Docker client
	mockClient := &mockDockerClient{}

	// Create mock LLM client
	mockLLM := &mockLLMClient{}

	// Create test facts
	testFacts := llm.Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "./mvnw clean package",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Env:       map[string]string{"SPRING_PROFILES_ACTIVE": "production"},
		Health:    "/actuator/health",
	}

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
	dockerfile, err := mockLLM.GenerateDockerfile(context.Background(), testFacts, "", "")
	if err != nil {
		t.Fatalf("Failed to generate Dockerfile: %v", err)
	}

	ok, logs, err := Verify(context.Background(), os.DirFS(tempDir), dockerfile, cfg.BuildTimeout)
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

	// Create mock dependencies
	dockerClient := &mockDockerClient{}
	f := llm.Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "./mvnw -B -ntp package -DskipTests",
		BuildDir:  ".",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Env:       map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
		Health:    "/actuator/health",
		BaseImage: "openjdk:11-jre-slim",
	}
	llmClient := &mockLLMClient{}

	// Create test config
	cfg := &config.Config{
		BuildTimeout: 15 * time.Minute,
		Debug:        true,
	}

	// Verify the Dockerfile
	ok, logs, err := Verify(ctx, os.DirFS(tmpDir), dockerfile, cfg.BuildTimeout)
	if err != nil {
		t.Fatalf("Failed to verify Dockerfile: %v", err)
	}

	// Check the output
	if !ok {
		t.Errorf("Expected verification to succeed, got logs: %s", logs)
	}

	// Verify build was attempted
	if dockerClient.buildCount == 0 {
		t.Error("Expected at least one build attempt")
	}
}
