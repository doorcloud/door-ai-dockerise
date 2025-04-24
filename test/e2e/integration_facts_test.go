//go:build integration
// +build integration

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
)

func TestE2E_FactsExtraction(t *testing.T) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.E2E {
		t.Skip("Skipping E2E test. Set DG_E2E=1 to run.")
	}

	if cfg.OpenAIKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Test cases
	tests := []struct {
		name      string
		testdata  string
		wantFacts facts.Facts
	}{
		{
			name:     "Spring Boot",
			testdata: "springboot",
			wantFacts: facts.Facts{
				Language:  "java",
				Framework: "spring-boot",
				BuildTool: "maven",
				BuildCmd:  "mvn clean package -DskipTests",
				BuildDir:  ".",
				StartCmd:  "java -jar app.jar",
				Artifact:  "target/*.jar",
				Ports:     []int{8080},
				Health:    "/actuator/health",
				BaseHint:  "eclipse-temurin:17-jdk",
				Env:       map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
		},
		{
			name:     "Node.js/Express",
			testdata: "nodeexpress",
			wantFacts: facts.Facts{
				Language:  "node",
				Framework: "express",
				BuildTool: "npm",
				BuildCmd:  "npm install",
				BuildDir:  ".",
				StartCmd:  "node dist/index.js",
				Artifact:  ".",
				Ports:     []int{3000},
				Health:    "/health",
				BaseHint:  "node:20-alpine",
				Env:       map[string]string{"NODE_ENV": "production"},
			},
		},
		{
			name:     "Python/Flask",
			testdata: "pythonflask",
			wantFacts: facts.Facts{
				Language:  "python",
				Framework: "flask",
				BuildTool: "pip",
				BuildCmd:  "pip install -r requirements.txt",
				BuildDir:  ".",
				StartCmd:  "python app.py",
				Artifact:  ".",
				Ports:     []int{5000},
				Health:    "/health",
				BaseHint:  "python:3.12-slim",
				Env:       map[string]string{"FLASK_ENV": "production"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path to testdata
			testdataPath := filepath.Join("..", "..", "testdata", tt.testdata)
			absPath, err := filepath.Abs(testdataPath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			// Verify testdata exists
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				t.Fatalf("Testdata not found at %s", absPath)
			}

			// Detect rule and extract facts
			rule, gotFacts, err := rules.DetectRule(absPath)
			if err != nil {
				t.Fatalf("DetectRule() error = %v", err)
			}

			if rule == nil {
				t.Fatal("DetectRule() returned nil rule")
			}

			// Convert rules.Facts to facts.Facts
			f := facts.Facts{
				Language:  gotFacts.Language,
				Framework: gotFacts.Framework,
				BuildTool: gotFacts.BuildTool,
				BuildCmd:  gotFacts.BuildCmd,
				BuildDir:  gotFacts.BuildDir,
				StartCmd:  gotFacts.StartCmd,
				Artifact:  gotFacts.Artifact,
				Ports:     gotFacts.Ports,
				Health:    gotFacts.Health,
				BaseHint:  gotFacts.BaseHint,
				Env:       make(map[string]string),
			}
			for k, v := range gotFacts.Env {
				f.Env[k] = v
			}

			// Verify all required facts fields
			if f.Language != tt.wantFacts.Language {
				t.Errorf("Language = %v, want %v", f.Language, tt.wantFacts.Language)
			}
			if f.Framework != tt.wantFacts.Framework {
				t.Errorf("Framework = %v, want %v", f.Framework, tt.wantFacts.Framework)
			}
			if f.BuildTool != tt.wantFacts.BuildTool {
				t.Errorf("BuildTool = %v, want %v", f.BuildTool, tt.wantFacts.BuildTool)
			}
			if f.BuildCmd != tt.wantFacts.BuildCmd {
				t.Errorf("BuildCmd = %v, want %v", f.BuildCmd, tt.wantFacts.BuildCmd)
			}
			if f.BuildDir != tt.wantFacts.BuildDir {
				t.Errorf("BuildDir = %v, want %v", f.BuildDir, tt.wantFacts.BuildDir)
			}
			if f.StartCmd != tt.wantFacts.StartCmd {
				t.Errorf("StartCmd = %v, want %v", f.StartCmd, tt.wantFacts.StartCmd)
			}
			if f.Artifact != tt.wantFacts.Artifact {
				t.Errorf("Artifact = %v, want %v", f.Artifact, tt.wantFacts.Artifact)
			}
			if len(f.Ports) != len(tt.wantFacts.Ports) {
				t.Errorf("Ports = %v, want %v", f.Ports, tt.wantFacts.Ports)
			} else {
				for i, port := range f.Ports {
					if port != tt.wantFacts.Ports[i] {
						t.Errorf("Ports[%d] = %v, want %v", i, port, tt.wantFacts.Ports[i])
					}
				}
			}
			if f.Health != tt.wantFacts.Health {
				t.Errorf("Health = %v, want %v", f.Health, tt.wantFacts.Health)
			}
			if f.BaseHint != tt.wantFacts.BaseHint {
				t.Errorf("BaseHint = %v, want %v", f.BaseHint, tt.wantFacts.BaseHint)
			}

			// Verify environment variables
			if len(f.Env) != len(tt.wantFacts.Env) {
				t.Errorf("Env = %v, want %v", f.Env, tt.wantFacts.Env)
			} else {
				for k, v := range tt.wantFacts.Env {
					if f.Env[k] != v {
						t.Errorf("Env[%s] = %v, want %v", k, f.Env[k], v)
					}
				}
			}

			// Validate facts
			if err := f.Validate(); err != nil {
				t.Errorf("Facts validation failed: %v", err)
			}

			// Validate build directory
			if err := facts.ValidateBuildDir(f, absPath); err != nil {
				t.Errorf("Build directory validation failed: %v", err)
			}
		})
	}
}
