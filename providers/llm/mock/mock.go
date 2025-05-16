package mock

import (
	"context"
	"errors"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type MockClient struct {
	responses map[string]string
}

func NewMockClient() *MockClient {
	return &MockClient{
		responses: make(map[string]string),
	}
}

func (m *MockClient) SetResponse(prompt string, response string) {
	m.responses[prompt] = response
}

func (m *MockClient) Complete(ctx context.Context, messages []core.Message) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("no messages provided")
	}

	// Check if this is a Spring Boot project
	for _, msg := range messages {
		if strings.Contains(msg.Content, "spring-boot") {
			return `FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn clean package -DskipTests

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
		}
	}

	// Check if this is a Maven project without wrapper
	if strings.Contains(messages[0].Content, "build_cmd: mvn ") {
		return `FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN mvn clean package

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
	}

	// Default response for non-Spring Boot projects
	return `FROM ubuntu:latest
WORKDIR /app
COPY . .
CMD ["./app"]`, nil
}

func (m *MockClient) Generate(ctx context.Context, facts core.Facts) (string, error) {
	if facts.StackType == "spring-boot" {
		return `FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn clean package -DskipTests

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
	}
	return "FROM ubuntu:latest\n", nil
}

func (m *MockClient) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	// If the error contains "ERROR", return a fixed version
	if strings.Contains(buildErr, "ERROR") {
		if strings.Contains(prevDockerfile, "spring-boot") {
			return `FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn clean package -DskipTests

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
		}
		return "FROM ubuntu:latest\nRUN apt-get update\n", nil
	}
	return prevDockerfile, nil
}

func (m *MockClient) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}, nil
}

func (m *MockClient) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	// Check if this is a Maven project without wrapper
	if strings.Contains(facts.StackType, "build_cmd: mvn ") {
		return `FROM maven:3.9-eclipse-temurin17 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn clean package -DskipTests

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]`, nil
	}

	// Default distroless builder for other cases
	return `FROM eclipse-temurin:17-jdk AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.m2 mvn -q package -DskipTests

# Runtime stage
FROM gcr.io/distroless/java17-debian12
WORKDIR /app
COPY --from=build /app/target/*.jar /app/app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/app.jar"]`, nil
}
