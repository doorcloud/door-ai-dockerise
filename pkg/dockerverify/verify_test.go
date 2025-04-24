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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockDockerClient struct {
	buildCount int
	containers map[string]bool
}

func (m *mockDockerClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	// Simulate that the image exists
	return []types.ImageSummary{
		{
			ID: "test-image-id",
			RepoTags: []string{
				"dockergen-e2e:test-digest",
			},
		},
	}, nil
}

func (m *mockDockerClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.buildCount++
	return types.ImageBuildResponse{
		Body: io.NopCloser(strings.NewReader("mock build output\nsha256:57ebb59e2dd7")),
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

var _ llm.Interface = (*mockLLMClient)(nil) // Verify interface implementation

func (m *mockLLMClient) GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error) {
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

func (m *mockLLMClient) FixDockerfile(ctx context.Context, facts map[string]interface{}, dockerfile string, buildDir string, errorLog string, errorType string, attempt int) (string, error) {
	// Return a fixed Dockerfile that uses mvnw
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

func (m *mockLLMClient) GenerateFacts(ctx context.Context, snippets []string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"language":   "java",
		"framework":  "spring-boot",
		"version":    "3.0.0",
		"build_tool": "maven",
		"build_cmd":  "./mvnw clean package",
		"artifact":   "target/*.jar",
		"ports":      []int{8080},
		"env":        map[string]string{"SPRING_PROFILES_ACTIVE": "production"},
		"health":     "/actuator/health",
	}, nil
}

func TestVerifyDockerfile_SkipRebuild(t *testing.T) {
	// Create a mock Docker client
	mockClient := &mockDockerClient{}

	// Create mock LLM client
	mockLLM := &mockLLMClient{}

	// Create test facts
	testFacts := facts.Facts{
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
	_, err := VerifyDockerfile(context.Background(), mockClient, tempDir, testFacts, mockLLM, 4, cfg)
	if err != nil {
		t.Fatalf("VerifyDockerfile failed: %v", err)
	}

	// Verify that we only built once
	if mockClient.buildCount > 1 {
		t.Errorf("Expected at most 1 build, got %d", mockClient.buildCount)
	}
}

func TestVerifyDockerfile(t *testing.T) {
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
	f := facts.Facts{
		Language:     "java",
		Framework:    "spring-boot",
		Version:      "2.7.0",
		BuildTool:    "maven",
		BuildCmd:     "./mvnw -B -ntp package -DskipTests",
		BuildDir:     ".",
		Artifact:     "target/*.jar",
		Ports:        []int{8080},
		Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
		Health:       "/actuator/health",
		Dependencies: []string{"spring-boot-starter-web"},
		BaseHint:     "openjdk",
	}
	llmClient := &mockLLMClient{}
	maxRetries := 3

	// Create test config
	cfg := &config.Config{
		BuildTimeout: 15 * time.Minute,
		Debug:        true,
	}

	// Verify the Dockerfile
	output, err := VerifyDockerfile(ctx, dockerClient, tmpDir, f, llmClient, maxRetries, cfg)
	if err != nil {
		t.Fatalf("Failed to verify Dockerfile: %v", err)
	}

	// Check the output
	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Verify build was attempted
	if dockerClient.buildCount == 0 {
		t.Error("Expected at least one build attempt")
	}
}
