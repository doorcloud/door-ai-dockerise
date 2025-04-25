package dockerfile

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aliou/dockerfile-gen/internal/facts"
	"github.com/sashabaranov/go-openai"
)

// mockOpenAIClient implements a mock OpenAI client for testing
type mockOpenAIClient struct {
	response string
	err      error
}

func (m *mockOpenAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if m.err != nil {
		return openai.ChatCompletionResponse{}, m.err
	}
	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: m.response,
				},
			},
		},
	}, nil
}

func TestGenerate(t *testing.T) {
	// Create test facts
	testFacts := facts.Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "./mvnw -q package -DskipTests",
		BuildDir:  ".",
		StartCmd:  "java -jar target/*.jar",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Health:    "/actuator/health",
		BaseImage: "eclipse-temurin:17-jdk",
		Env:       map[string]string{},
	}

	// Set up mock client
	mockClient := &mockOpenAIClient{
		response: `FROM eclipse-temurin:17-jdk
WORKDIR /app
COPY mvnw .
COPY .mvn .mvn
COPY pom.xml .
RUN ./mvnw -q package -DskipTests
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/actuator/health || exit 1
CMD ["java", "-jar", "target/*.jar"]`,
	}

	// Override OpenAI client creation
	origNewClient := openai.NewClient
	openai.NewClient = func(apiKey string) *openai.Client {
		return mockClient
	}
	defer func() {
		openai.NewClient = origNewClient
	}()

	// Generate Dockerfile
	dockerfile, err := Generate(context.Background(), testFacts, "", "")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check required commands
	requiredCommands := []string{
		"FROM",
		"WORKDIR",
		"COPY",
		"RUN",
		"EXPOSE",
		"HEALTHCHECK",
		"CMD",
	}

	for _, cmd := range requiredCommands {
		if !strings.Contains(dockerfile, cmd) {
			t.Errorf("Dockerfile missing command: %s", cmd)
		}
	}

	// Check specific values
	if !strings.Contains(dockerfile, testFacts.BaseImage) {
		t.Error("Dockerfile missing base image")
	}
	if !strings.Contains(dockerfile, testFacts.BuildCmd) {
		t.Error("Dockerfile missing build command")
	}
	if !strings.Contains(dockerfile, testFacts.StartCmd) {
		t.Error("Dockerfile missing start command")
	}
	if !strings.Contains(dockerfile, testFacts.Health) {
		t.Error("Dockerfile missing health check")
	}

	// Test error case
	t.Run("openai error", func(t *testing.T) {
		mockClient.err = fmt.Errorf("mock error")
		_, err := Generate(context.Background(), testFacts, "", "")
		if err == nil {
			t.Error("Generate() expected error, got nil")
		}
	})
}
