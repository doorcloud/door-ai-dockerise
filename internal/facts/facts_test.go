package facts

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

// mockClient implements the llm.Client interface for testing
type mockClient struct{}

func (m *mockClient) Chat(model, prompt string) (string, error) {
	return `{
		"language": "java",
		"framework": "spring-boot",
		"build_tool": "maven",
		"build_cmd": "./mvnw -q package",
		"build_dir": ".",
		"start_cmd": "java -jar target/*.jar",
		"artifact": "target/*.jar",
		"ports": [8080],
		"health": "/actuator/health",
		"env": {},
		"base_image": "eclipse-temurin:17-jdk"
	}`, nil
}

func TestInfer(t *testing.T) {
	// Create test filesystem
	fsys := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.7.0</version>
    </parent>
</project>`),
		},
		"mvnw": &fstest.MapFile{
			Data: []byte("#!/bin/sh"),
		},
		".mvn/wrapper/maven-wrapper.jar": &fstest.MapFile{
			Data: []byte("mock jar"),
		},
	}

	// Test inference with mock client
	rule := detect.Rule{
		Name: "spring-boot",
		Tool: "maven",
	}

	facts, err := InferWithClient(context.Background(), fsys, rule, &mockClient{})
	if err != nil {
		t.Fatalf("Infer() error = %v", err)
	}

	// Verify inferred facts
	if facts.Language != "java" {
		t.Errorf("Language = %v, want java", facts.Language)
	}
	if facts.Framework != "spring-boot" {
		t.Errorf("Framework = %v, want spring-boot", facts.Framework)
	}
	if facts.BuildTool != "maven" {
		t.Errorf("BuildTool = %v, want maven", facts.BuildTool)
	}
	if facts.BuildCmd != "./mvnw -q package" {
		t.Errorf("BuildCmd = %v, want ./mvnw -q package", facts.BuildCmd)
	}
	if facts.StartCmd != "java -jar target/*.jar" {
		t.Errorf("StartCmd = %v, want java -jar target/*.jar", facts.StartCmd)
	}
	if len(facts.Ports) != 1 || facts.Ports[0] != 8080 {
		t.Errorf("Ports = %v, want [8080]", facts.Ports)
	}
	if facts.Health != "/actuator/health" {
		t.Errorf("Health = %v, want /actuator/health", facts.Health)
	}
	if facts.BaseImage != "eclipse-temurin:17-jdk" {
		t.Errorf("BaseImage = %v, want eclipse-temurin:17-jdk", facts.BaseImage)
	}
}
