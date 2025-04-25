package rules

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

func TestDetectStack(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantRule *detect.Rule
		wantErr  error
	}{
		{
			name: "spring boot maven",
			files: map[string]string{
				"pom.xml": "<project></project>",
			},
			wantRule: &detect.Rule{
				Name: "spring-boot",
				Tool: "maven",
			},
			wantErr: nil,
		},
		{
			name: "spring boot gradle",
			files: map[string]string{
				"gradlew": "#!/bin/sh",
			},
			wantRule: &detect.Rule{
				Name: "spring-boot",
				Tool: "gradle",
			},
			wantErr: nil,
		},
		{
			name:     "no match",
			files:    map[string]string{},
			wantRule: nil,
			wantErr:  ErrUnknownStack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test filesystem
			fsys := fstest.MapFS{}
			for path, content := range tt.files {
				fsys[path] = &fstest.MapFile{Data: []byte(content)}
			}

			// Run detection
			gotRule, gotErr := DetectStack(fsys)
			if gotErr != tt.wantErr {
				t.Errorf("DetectStack() error = %v, want %v", gotErr, tt.wantErr)
			}
			if tt.wantRule != nil && *gotRule != *tt.wantRule {
				t.Errorf("DetectStack() rule = %v, want %v", gotRule, tt.wantRule)
			}
		})
	}
}

func TestDetect_SpringBoot(t *testing.T) {
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
	rule := Detect(dir)
	if rule == nil {
		t.Error("Expected Spring Boot rule to match")
	}
	if rule.Name() != "spring-boot" {
		t.Errorf("Expected rule name 'spring-boot', got '%s'", rule.Name())
	}
}

func TestDetect_EmptyDir(t *testing.T) {
	// Create a temporary empty directory
	dir, err := os.MkdirTemp("", "empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Test detection
	rule := Detect(dir)
	if rule != nil {
		t.Error("Expected no rule to match empty directory")
	}
}
