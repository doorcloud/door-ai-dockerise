package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/mock"
	"github.com/stretchr/testify/require"
)

func TestDockerfile_NoWrapper(t *testing.T) {
	mockClient := mock.NewMockClient()
	generator := generate.NewLLM(mockClient)

	facts := core.Facts{
		StackType: "spring-boot",
		BuildTool: "maven",
		Port:      8080,
	}

	dockerfile, err := generator.GenerateDockerfile(context.Background(), facts)
	require.NoError(t, err)

	// Check for builder image
	require.Contains(t, dockerfile, "FROM eclipse-temurin:17-jdk AS build")

	// Check for distroless base image
	require.Contains(t, dockerfile, "gcr.io/distroless/java17-debian12")

	// Check for cache mount
	require.Contains(t, dockerfile, "--mount=type=cache,target=/root/.m2")
}

// mockGenerator implements core.Generator for testing
type mockGenerator struct{}

func (g *mockGenerator) Generate(ctx context.Context, facts core.Facts) (string, error) {
	if strings.Contains(facts.BuildTool, "maven") {
		return `FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn clean package

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
	}
	return "", nil
}

func (g *mockGenerator) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	return "", nil
}
