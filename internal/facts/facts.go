package facts

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/util"
)

// LLMClient is an interface for LLM operations.
type LLMClient interface {
	GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error)
	FixDockerfile(ctx context.Context, facts map[string]interface{}, dockerfile string, buildDir string, errorLog string, errorType string, attempt int) (string, error)
	GenerateFacts(ctx context.Context, snippets []string) (map[string]interface{}, error)
}

// Facts represents the extracted information about a project.
type Facts struct {
	// Language is the primary programming language (e.g., "java", "javascript").
	Language string `json:"language,omitempty"`

	// Framework is the web framework (e.g., "spring-boot", "express").
	Framework string `json:"framework,omitempty"`

	// Version is the framework version.
	Version string `json:"version,omitempty"`

	// BuildTool is the build system (e.g., "maven", "npm").
	BuildTool string `json:"build_tool,omitempty"`

	// BuildCmd is the command to build the project.
	BuildCmd string `json:"build_cmd"`

	// BuildDir is the directory containing the build manifest (e.g., pom.xml, package.json).
	BuildDir string `json:"build_dir,omitempty"`

	// StartCmd is the command to start the application.
	StartCmd string `json:"start_cmd,omitempty"`

	// Artifact is the path to the built artifact.
	Artifact string `json:"artifact,omitempty"`

	// Ports are the exposed ports.
	Ports []int `json:"ports,omitempty"`

	// Env are environment variables.
	Env map[string]string `json:"env,omitempty"`

	// Health is the health check endpoint.
	Health string `json:"health,omitempty"`

	// Dependencies are the project dependencies.
	Dependencies []string `json:"dependencies,omitempty"`

	// BaseHint is a hint for base image selection
	BaseHint string `json:"base_hint,omitempty"`
}

// Validate checks if the facts are complete and valid.
func (f Facts) Validate() error {
	var missing []string

	if f.Language == "" {
		missing = append(missing, "language")
	}
	if f.Framework == "" {
		missing = append(missing, "framework")
	}
	if f.BuildTool == "" {
		missing = append(missing, "build_tool")
	}
	if f.BuildCmd == "" {
		missing = append(missing, "build_cmd")
	}
	if f.BuildDir == "" {
		missing = append(missing, "build_dir")
	}
	if f.StartCmd == "" {
		missing = append(missing, "start_cmd")
	}
	if f.Artifact == "" {
		missing = append(missing, "artifact")
	}
	if len(f.Ports) == 0 {
		missing = append(missing, "ports")
	}
	if len(f.Env) == 0 {
		missing = append(missing, "env")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required facts: %s", strings.Join(missing, ", "))
	}

	// Check Spring/Maven specific requirements
	if !strings.Contains(f.BuildCmd, "-f") {
		// Check both build directory and root
		if !util.FileExists(filepath.Join(f.BuildDir, "pom.xml")) && !util.FileExists("pom.xml") {
			return fmt.Errorf("build cmd uses Maven but no pom.xml found in %s or root", f.BuildDir)
		}
	}

	return nil
}

// FromJSON parses facts from JSON.
func FromJSON(data interface{}) (Facts, error) {
	var f Facts

	// If data is already a map, marshal it to JSON first
	if m, ok := data.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(m)
		if err != nil {
			return Facts{}, fmt.Errorf("marshal facts map: %w", err)
		}
		data = jsonData
	}

	// Parse JSON bytes
	if jsonData, ok := data.([]byte); ok {
		if err := json.Unmarshal(jsonData, &f); err != nil {
			return Facts{}, fmt.Errorf("parse facts: %w", err)
		}
		return f, nil
	}

	return Facts{}, fmt.Errorf("unsupported facts data type")
}

// ToJSON converts facts to JSON.
func (f Facts) ToJSON() ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// ToMap converts facts to a map.
func (f Facts) ToMap() map[string]interface{} {
	// Marshal to JSON then unmarshal to map to handle all fields
	data, _ := f.ToJSON()
	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)
	return m
}

// Log writes the facts to the logger.
func (f Facts) Log(logger *slog.Logger) {
	if os.Getenv("DG_DEBUG") != "1" {
		return
	}

	logger.Debug("project facts",
		"language", f.Language,
		"framework", f.Framework,
		"version", f.Version,
		"build_tool", f.BuildTool,
		"build_cmd", f.BuildCmd,
		"build_dir", f.BuildDir,
		"artifact", f.Artifact,
		"ports", f.Ports,
		"env", f.Env,
		"health", f.Health,
		"dependencies", f.Dependencies,
	)
}
