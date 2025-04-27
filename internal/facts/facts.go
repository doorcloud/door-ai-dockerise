package facts

import (
	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

// Facts represents the structured information about an application
type Facts struct {
	// Language information
	Language        string `json:"language"`
	LanguageVersion string `json:"language_version"`

	// Framework information
	Framework        string `json:"framework"`
	FrameworkVersion string `json:"framework_version"`

	// Build information
	BuildTool    string `json:"build_tool"`
	BuildCommand string `json:"build_command"`
	Artifact     string `json:"artifact"`

	// Runtime information
	Ports       []int             `json:"ports"`
	HealthCheck string            `json:"health_check"`
	Environment map[string]string `json:"environment"`

	// Dependencies
	Dependencies []string `json:"dependencies"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// Generator interface for generating Facts from detection results
type Generator interface {
	Generate(detectResult detect.Result) (Facts, error)
}

// NewFacts creates a new Facts struct with default values
func NewFacts() Facts {
	return Facts{
		Environment: make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
}
