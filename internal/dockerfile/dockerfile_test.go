package dockerfile

import (
	"context"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

func TestGenerate(t *testing.T) {
	// Create test facts
	testFacts := types.Facts{
		Language:  "java",
		Framework: "spring-boot",
		BuildTool: "maven",
		BuildCmd:  "./mvnw -q package -DskipTests",
		BuildDir:  ".",
		StartCmd:  "java -jar target/*.jar",
		Artifact:  "target/*.jar",
		Ports:     []int{8080},
		Health:    "/actuator/health",
		BaseImage: "openjdk:11-jdk",
		Env:       map[string]string{},
	}

	// Generate Dockerfile
	dockerfile, err := Generate(context.Background(), testFacts, &llm.MockClient{})
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
	if !strings.Contains(dockerfile, strings.TrimPrefix(testFacts.BuildCmd, "./")) {
		t.Error("Dockerfile missing build command")
	}
	if !strings.Contains(dockerfile, `CMD ["java", "-jar", "target/*.jar"]`) {
		t.Error("Dockerfile missing start command")
	}
	if !strings.Contains(dockerfile, testFacts.Health) {
		t.Error("Dockerfile missing health check")
	}
}
