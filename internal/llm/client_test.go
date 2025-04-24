package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockClient struct{}

func (m *mockClient) AnalyzeFacts(ctx context.Context, snippets []string) (Facts, error) {
	return Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "./mvnw -q package -DskipTests",
		StartCmd:  "java -jar target/*.jar",
		Ports:     []int{8080},
		Health:    "/actuator/health",
	}, nil
}

func (m *mockClient) GenerateDockerfile(ctx context.Context, facts Facts, prevDockerfile string, prevError string) (string, error) {
	return `FROM eclipse-temurin:17-jdk
WORKDIR /app
COPY . .
RUN ./mvnw -q package -DskipTests
CMD ["java", "-jar", "target/*.jar"]`, nil
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	client := &mockClient{}

	// Test AnalyzeFacts
	facts, err := client.AnalyzeFacts(ctx, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "spring-boot", facts.Framework)
	assert.Equal(t, "maven", facts.BuildTool)
	assert.Equal(t, "./mvnw -q package -DskipTests", facts.BuildCmd)
	assert.Equal(t, "java -jar target/*.jar", facts.StartCmd)
	assert.Equal(t, []int{8080}, facts.Ports)
	assert.Equal(t, "/actuator/health", facts.Health)

	// Test GenerateDockerfile
	dockerfile, err := client.GenerateDockerfile(ctx, facts, "", "")
	assert.NoError(t, err)
	assert.Contains(t, dockerfile, "FROM eclipse-temurin:17-jdk")
	assert.Contains(t, dockerfile, "WORKDIR /app")
	assert.Contains(t, dockerfile, "COPY . .")
	assert.Contains(t, dockerfile, "RUN ./mvnw -q package -DskipTests")
	assert.Contains(t, dockerfile, `CMD ["java", "-jar", "target/*.jar"]`)
}
