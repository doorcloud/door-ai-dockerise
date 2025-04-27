package llm

import (
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
)

// BasicSummarizer implements the Summarizer interface
type BasicSummarizer struct{}

// NewBasicSummarizer creates a new BasicSummarizer
func NewBasicSummarizer() *BasicSummarizer {
	return &BasicSummarizer{}
}

// Summarize converts Facts into an ApplicationSpec
func (s *BasicSummarizer) Summarize(facts facts.Facts) (ApplicationSpec, error) {
	spec := NewApplicationSpec()

	// Set basic application information
	spec.Name = facts.Framework
	spec.Type = "web" // Default to web application
	spec.Description = "A " + facts.Framework + " application"

	// Set runtime configuration
	spec.Runtime.BaseImage = facts.Language + ":" + facts.LanguageVersion
	spec.Runtime.Ports = facts.Ports
	spec.Runtime.HealthCheck = facts.HealthCheck
	for k, v := range facts.Environment {
		spec.Runtime.Environment[k] = v
	}

	// Set build configuration
	spec.Build.Context = "."
	spec.Build.Dockerfile = "Dockerfile"
	if facts.BuildTool != "" {
		spec.Build.Args["BUILD_TOOL"] = facts.BuildTool
	}

	// Set dependencies
	spec.Dependencies = facts.Dependencies

	// Copy metadata
	for k, v := range facts.Metadata {
		spec.Metadata[k] = v
	}

	return spec, nil
}
