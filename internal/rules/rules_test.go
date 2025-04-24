package rules

import (
	"path/filepath"
	"testing"
)

func TestRuleDetection(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		wantRule  string
		wantFacts Facts
	}{
		{
			name:     "Spring Boot",
			repoPath: "../testdata/springboot",
			wantRule: "springboot",
			wantFacts: Facts{
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
			repoPath: "../testdata/nodeexpress",
			wantRule: "nodeexpress",
			wantFacts: Facts{
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
			repoPath: "../testdata/pythonflask",
			wantRule: "pythonflask",
			wantFacts: Facts{
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
			absPath, err := filepath.Abs(tt.repoPath)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}

			rule, facts, err := DetectRule(absPath)
			if err != nil {
				t.Fatalf("DetectRule() error = %v", err)
			}

			if rule == nil {
				t.Fatal("DetectRule() returned nil rule")
			}

			if facts.Language != tt.wantFacts.Language {
				t.Errorf("Language = %v, want %v", facts.Language, tt.wantFacts.Language)
			}
			if facts.Framework != tt.wantFacts.Framework {
				t.Errorf("Framework = %v, want %v", facts.Framework, tt.wantFacts.Framework)
			}
			if facts.BuildTool != tt.wantFacts.BuildTool {
				t.Errorf("BuildTool = %v, want %v", facts.BuildTool, tt.wantFacts.BuildTool)
			}
			if facts.BuildCmd != tt.wantFacts.BuildCmd {
				t.Errorf("BuildCmd = %v, want %v", facts.BuildCmd, tt.wantFacts.BuildCmd)
			}
			if facts.BuildDir != tt.wantFacts.BuildDir {
				t.Errorf("BuildDir = %v, want %v", facts.BuildDir, tt.wantFacts.BuildDir)
			}
			if facts.StartCmd != tt.wantFacts.StartCmd {
				t.Errorf("StartCmd = %v, want %v", facts.StartCmd, tt.wantFacts.StartCmd)
			}
			if facts.Artifact != tt.wantFacts.Artifact {
				t.Errorf("Artifact = %v, want %v", facts.Artifact, tt.wantFacts.Artifact)
			}
			if len(facts.Ports) != len(tt.wantFacts.Ports) {
				t.Errorf("Ports = %v, want %v", facts.Ports, tt.wantFacts.Ports)
			} else {
				for i, port := range facts.Ports {
					if port != tt.wantFacts.Ports[i] {
						t.Errorf("Ports[%d] = %v, want %v", i, port, tt.wantFacts.Ports[i])
					}
				}
			}
			if facts.Health != tt.wantFacts.Health {
				t.Errorf("Health = %v, want %v", facts.Health, tt.wantFacts.Health)
			}
			if facts.BaseHint != tt.wantFacts.BaseHint {
				t.Errorf("BaseHint = %v, want %v", facts.BaseHint, tt.wantFacts.BaseHint)
			}
		})
	}
}
