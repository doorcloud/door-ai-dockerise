package facts

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/stretchr/testify/assert"
)

func TestInferWithClient(t *testing.T) {
	// Create a mock filesystem with a Java Spring Boot application
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`
				<project>
					<groupId>com.example</groupId>
					<artifactId>demo</artifactId>
					<version>0.0.1-SNAPSHOT</version>
					<dependencies>
						<dependency>
							<groupId>org.springframework.boot</groupId>
							<artifactId>spring-boot-starter-web</artifactId>
						</dependency>
					</dependencies>
				</project>
			`),
		},
	}

	// Create fixture directory and file
	fixtureDir := filepath.Join("testdata", "fixtures", "facts")
	if err := os.MkdirAll(fixtureDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("testdata")

	fixturePath := filepath.Join(fixtureDir, "response.json")
	fixtureContent := `{
		"choices": [
			{
				"message": {
					"content": "{\"language\":\"java\",\"framework\":\"spring-boot\",\"build_tool\":\"maven\",\"build_cmd\":\"mvn clean package\",\"build_dir\":\"target\",\"start_cmd\":\"java -jar target/demo-0.0.1-SNAPSHOT.jar\",\"artifact\":\"target/demo-0.0.1-SNAPSHOT.jar\",\"ports\":[8080],\"health\":\"/actuator/health\",\"base_image\":\"eclipse-temurin:17-jre\",\"env\":{\"SPRING_PROFILES_ACTIVE\":\"prod\"}}"
				}
			}
		]
	}`
	if err := os.WriteFile(fixturePath, []byte(fixtureContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a mock client
	mockClient := &llm.MockClient{}

	// Create a mock rule info
	rule := detect.RuleInfo{
		Name: "spring-boot",
		Tool: "maven",
	}

	// Call InferWithClient
	facts, err := InferWithClient(context.Background(), fsys, rule, mockClient)
	assert.NoError(t, err)

	// Verify the facts
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "spring-boot", facts.Framework)
	assert.Equal(t, "maven", facts.BuildTool)
	assert.Equal(t, "mvn clean package", facts.BuildCmd)
	assert.Equal(t, "target", facts.BuildDir)
	assert.Equal(t, "java -jar target/demo-0.0.1-SNAPSHOT.jar", facts.StartCmd)
	assert.Equal(t, "target/demo-0.0.1-SNAPSHOT.jar", facts.Artifact)
	assert.Equal(t, []int{8080}, facts.Ports)
	assert.Equal(t, "/actuator/health", facts.Health)
	assert.Equal(t, "eclipse-temurin:17-jre", facts.BaseImage)
	assert.Equal(t, map[string]string{"SPRING_PROFILES_ACTIVE": "prod"}, facts.Env)
}

func TestGetFacts(t *testing.T) {
	// Create a mock filesystem with a pom.xml file
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`
				<project>
					<groupId>com.example</groupId>
					<artifactId>demo</artifactId>
					<version>0.0.1-SNAPSHOT</version>
					<dependencies>
						<dependency>
							<groupId>org.springframework.boot</groupId>
							<artifactId>spring-boot-starter-web</artifactId>
						</dependency>
					</dependencies>
				</project>
			`),
		},
	}

	// Create a mock rule info
	rule := detect.RuleInfo{
		Name: "spring-boot",
		Tool: "maven",
	}

	// Call GetFactsFromRule
	facts, err := GetFactsFromRule(fsys, rule)
	assert.NoError(t, err)

	// Verify the facts
	assert.Equal(t, "java", facts.Language)
	assert.Equal(t, "spring-boot", facts.Framework)
	assert.Equal(t, "maven", facts.BuildTool)
	assert.Equal(t, "mvn clean package", facts.BuildCmd)
	assert.Equal(t, "target", facts.BuildDir)
	assert.Equal(t, "java -jar target/demo-0.0.1-SNAPSHOT.jar", facts.StartCmd)
	assert.Equal(t, "target/demo-0.0.1-SNAPSHOT.jar", facts.Artifact)
	assert.Equal(t, []int{8080}, facts.Ports)
	assert.Equal(t, "/actuator/health", facts.Health)
	assert.Equal(t, "eclipse-temurin:17-jre", facts.BaseImage)
	assert.Equal(t, map[string]string{"SPRING_PROFILES_ACTIVE": "prod"}, facts.Env)
}
