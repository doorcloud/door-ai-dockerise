package loop

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
)

// mockClient implements the llm.Client interface for testing
type mockClient struct{}

func (m *mockClient) Chat(prompt string, model string) (string, error) {
	switch model {
	case "facts":
		return `{
			"language": "java",
			"framework": "spring-boot",
			"build_tool": "maven",
			"build_cmd": "./mvnw -q package -DskipTests",
			"build_dir": ".",
			"start_cmd": "java -jar target/*.jar",
			"artifact": "target/*.jar",
			"ports": [8080],
			"health": "/actuator/health",
			"env": {},
			"base_image": "openjdk:11-jdk"
		}`, nil
	case "dockerfile":
		return `FROM openjdk:11-jdk
WORKDIR /app
COPY mvnw .
COPY .mvn .mvn
COPY pom.xml .
RUN chmod +x ./mvnw && ./mvnw -q package -DskipTests
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080/actuator/health || exit 1
CMD ["java", "-jar", "target/*.jar"]`, nil
	default:
		return "", nil
	}
}

func TestRun(t *testing.T) {
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`<project></project>`),
		},
	}

	client := &llm.MockClient{}
	dockerfile, err := Run(context.Background(), fsys, client)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if dockerfile == "" {
		t.Error("Run() returned empty dockerfile")
	}
}
