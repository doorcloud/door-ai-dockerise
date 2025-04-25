package springboot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpringBootRule(t *testing.T) {
	// Create a temporary directory for testing
	dir, err := os.MkdirTemp("", "springboot-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a Spring Boot application class
	appDir := filepath.Join(dir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	appFile := filepath.Join(appDir, "Application.java")
	if err := os.WriteFile(appFile, []byte(`
package com.example;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	rule := &springBootRule{}
	if !rule.Detect(dir) {
		t.Error("Expected Spring Boot rule to match")
	}

	// Test facts
	facts := rule.Facts(dir)
	if facts["framework"] != "Spring Boot" {
		t.Errorf("Expected framework 'Spring Boot', got '%s'", facts["framework"])
	}
	if facts["language"] != "Java" {
		t.Errorf("Expected language 'Java', got '%s'", facts["language"])
	}
}
