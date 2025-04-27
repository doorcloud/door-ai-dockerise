package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts/spring"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/ollama"
)

// Option is a function that configures a Pipeline
type Option func(*Pipeline)

// WithDetectors sets the detectors for the pipeline
func WithDetectors(detectors ...detectors.Detector) Option {
	return func(p *Pipeline) {
		p.detectors = detectors
	}
}

// WithFactProviders sets the fact providers for the pipeline
func WithFactProviders(providers ...facts.Provider) Option {
	return func(p *Pipeline) {
		p.factProviders = providers
	}
}

// WithGenerator sets the generator for the pipeline
func WithGenerator(generator generate.Generator) Option {
	return func(p *Pipeline) {
		p.generator = generator
	}
}

// WithDockerDriver sets the docker driver for the pipeline
func WithDockerDriver(driver docker.Driver) Option {
	return func(p *Pipeline) {
		p.dockerDriver = driver
	}
}

// ErrNoDetectorMatch is returned when no detector matches the project
var ErrNoDetectorMatch = errors.New("no detector matched the project type")

type Pipeline struct {
	detectors     []detectors.Detector
	factProviders []facts.Provider
	generator     generate.Generator
	dockerDriver  docker.Driver
	cfg           *config.Config
}

func New(cfg *config.Config, opts ...Option) *Pipeline {
	p := &Pipeline{
		cfg: cfg,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Pipeline) Run(ctx context.Context, projectDir string) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.cfg.DockerTimeout)
	defer cancel()

	// Detect project type
	fsys := os.DirFS(projectDir)
	var stackInfo core.StackInfo
	var detected bool
	for _, detector := range p.detectors {
		info, found, err := detector.Detect(ctx, fsys, nil)
		if err != nil {
			return err
		}
		if found {
			stackInfo = info
			detected = true
			break
		}
	}
	if !detected {
		return ErrNoDetectorMatch
	}

	// When stack is spring-boot, extract facts and use them in the prompt
	var prompt string
	if stackInfo.Name == "spring-boot" {
		spec, err := spring.NewExtractor().Extract(projectDir)
		if err != nil {
			// Fall back to current path if extractor errors
			spec = nil
		}

		// Use the spec in the prompt
		prompt = fmt.Sprintf("Generate a Dockerfile for a Spring Boot application with the following configuration:\n"+
			"Build Tool: %s\n"+
			"JDK Version: %s\n"+
			"Spring Boot Version: %s\n"+
			"Build Command: %s\n"+
			"Artifact: %s\n"+
			"Health Endpoint: %s\n"+
			"Ports: %v\n",
			spec.BuildTool,
			spec.JDKVersion,
			spec.SpringBootVersion,
			spec.BuildCmd,
			spec.Artifact,
			spec.HealthEndpoint,
			spec.Ports)
	} else {
		prompt = "Generate a Dockerfile for the project."
	}

	// Use the prompt in the chat completion
	llm := ollama.New()
	messages := []core.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	response, err := llm.Complete(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Write the Dockerfile
	if err := os.WriteFile(filepath.Join(projectDir, "Dockerfile"), []byte(response), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Gather facts
	facts := make(map[string]interface{})
	for _, provider := range p.factProviders {
		providerFacts, err := provider.Gather(projectDir)
		if err != nil {
			return err
		}
		for k, v := range providerFacts {
			facts[k] = v
		}
	}

	// Convert facts to core.Facts
	coreFacts := core.Facts{
		StackType: stackInfo.Name,
		BuildTool: stackInfo.BuildTool,
		Port:      stackInfo.Port,
	}

	// Generate Dockerfile
	dockerfile, err := p.generator.Generate(ctx, coreFacts)
	if err != nil {
		return err
	}

	// Verify Dockerfile
	if err := p.dockerDriver.Verify(ctx, dockerfile); err != nil {
		return err
	}

	return nil
}
