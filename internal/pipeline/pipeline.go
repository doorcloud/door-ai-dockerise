package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts/spring"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	dockerverifier "github.com/doorcloud/door-ai-dockerise/adapters/verifiers/docker"
	"github.com/doorcloud/door-ai-dockerise/core"
	dockerdriver "github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/doorcloud/door-ai-dockerise/providers/llm/ollama"
)

// Option is a function that configures a Pipeline
type Option func(*Pipeline)

// WithDetectors sets the detectors for the pipeline
func WithDetectors(detectors ...core.Detector) Option {
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
func WithDockerDriver(driver dockerdriver.Driver) Option {
	return func(p *Pipeline) {
		p.dockerDriver = driver
	}
}

// ErrNoDetectorMatch is returned when no detector matches the project
var ErrNoDetectorMatch = errors.New("no detector matched the project type")

type Pipeline struct {
	detectors     []core.Detector
	factProviders []facts.Provider
	generator     generate.Generator
	dockerDriver  dockerdriver.Driver
	cfg           *config.Config
	confidenceMin float64
}

func New(cfg *config.Config, opts ...Option) *Pipeline {
	p := &Pipeline{
		cfg:           cfg,
		confidenceMin: 0.5, // Default minimum confidence
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Pipeline) SetConfidenceThreshold(threshold float64) {
	p.confidenceMin = threshold
}

func (p *Pipeline) Run(ctx context.Context, projectDir string, logs io.Writer) error {
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

	if stackInfo.Confidence < p.confidenceMin {
		return fmt.Errorf("detection confidence %f below threshold %f", stackInfo.Confidence, p.confidenceMin)
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
	if err := os.WriteFile(filepath.Join(projectDir, "Dockerfile"), []byte(response), 0o644); err != nil {
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
	if _, err := p.generator.Generate(ctx, coreFacts); err != nil {
		return err
	}

	// Build image
	opts := dockerdriver.BuildOptions{
		Context:    projectDir,
		Dockerfile: "Dockerfile",
		Tags:       []string{"myapp:latest"},
	}
	if err := p.dockerDriver.Build(ctx, "Dockerfile", opts); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Verify container health
	if stackInfo.Name == "spring-boot" {
		spec, err := spring.NewExtractor().Extract(projectDir)
		if err != nil {
			return fmt.Errorf("failed to extract Spring Boot facts: %w", err)
		}

		// Create a buffer to capture build and health check logs
		var logBuffer strings.Builder

		// Create health verifier
		verifier := dockerverifier.NewRunner(p.dockerDriver)

		// Run health check with timeout
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Run container and verify health
		if err := verifier.Run(ctx, "myapp:latest", spec, &logBuffer); err != nil {
			return fmt.Errorf("health check failed: %w\n%s", err, logBuffer.String())
		}
	}

	return nil
}
