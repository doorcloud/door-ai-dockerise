package facts

import (
	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

// BasicGenerator implements the Generator interface
type BasicGenerator struct{}

// NewBasicGenerator creates a new BasicGenerator
func NewBasicGenerator() *BasicGenerator {
	return &BasicGenerator{}
}

// Generate converts a detect.Result into Facts
func (g *BasicGenerator) Generate(detectResult detect.Result) (Facts, error) {
	facts := NewFacts()

	// Map stack information to language/framework
	facts.Language = detectResult.StackName
	facts.LanguageVersion = detectResult.StackVersion

	// Copy build information
	facts.BuildTool = detectResult.BuildTool
	facts.BuildCommand = detectResult.BuildCommand
	facts.Artifact = detectResult.Artifact

	// Copy runtime information
	facts.Ports = detectResult.Ports
	facts.HealthCheck = detectResult.HealthCheck
	for k, v := range detectResult.Environment {
		facts.Environment[k] = v
	}

	// Copy dependencies
	facts.Dependencies = detectResult.Dependencies

	// Copy metadata
	for k, v := range detectResult.Metadata {
		facts.Metadata[k] = v
	}

	return facts, nil
}
