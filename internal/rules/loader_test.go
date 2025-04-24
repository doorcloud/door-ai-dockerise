package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRules(t *testing.T) {
	// Create a temporary directory for test rules
	tempDir, err := os.MkdirTemp("", "rules-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test rule file
	ruleContent := `kind: stackRule/v1
id: spring-boot
detect:
  globs:
    - "**/pom.xml"
    - "**/build.gradle*"
  containsRegex: "@SpringBootApplication"
hints:
  language: java
  framework: spring-boot
  build:
    tool: maven
    cmd: "./mvnw -q package -DskipTests"
    dir: .
  ports: [8080]
  health: "/actuator/health"
  baseImage: "eclipse-temurin:17-jdk"`

	rulePath := filepath.Join(tempDir, "springboot.yaml")
	if err := os.WriteFile(rulePath, []byte(ruleContent), 0644); err != nil {
		t.Fatalf("Failed to write rule file: %v", err)
	}

	// Load rules
	rules, err := LoadRules(tempDir)
	if err != nil {
		t.Fatalf("Failed to load rules: %v", err)
	}

	// Verify loaded rules
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	if rule.Kind != "stackRule/v1" {
		t.Errorf("Expected kind 'stackRule/v1', got '%s'", rule.Kind)
	}
	if rule.ID != "spring-boot" {
		t.Errorf("Expected ID 'spring-boot', got '%s'", rule.ID)
	}
	if len(rule.Detect.Globs) != 2 {
		t.Errorf("Expected 2 globs, got %d", len(rule.Detect.Globs))
	}
	if rule.Hints.Language != "java" {
		t.Errorf("Expected language 'java', got '%s'", rule.Hints.Language)
	}
	if rule.Hints.Framework != "spring-boot" {
		t.Errorf("Expected framework 'spring-boot', got '%s'", rule.Hints.Framework)
	}
	if rule.Hints.Build.Tool != "maven" {
		t.Errorf("Expected build tool 'maven', got '%s'", rule.Hints.Build.Tool)
	}
	if len(rule.Hints.Ports) != 1 || rule.Hints.Ports[0] != 8080 {
		t.Errorf("Expected ports [8080], got %v", rule.Hints.Ports)
	}
}
