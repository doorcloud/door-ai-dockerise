package dockerfile

import (
	"context"
	"strings"
	"testing"

	"github.com/aliou/dockerfile-gen/internal/facts"
)

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

	// Generate Dockerfile
	dockerfile, err := Generate(context.Background(), testFacts)
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
}
