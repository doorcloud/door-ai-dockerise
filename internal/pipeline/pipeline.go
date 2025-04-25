package pipeline

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// Pipeline represents the Dockerfile generation pipeline
type Pipeline struct {
	fsys fs.FS
	reg  *rules.Registry
}

// New creates a new pipeline for the given directory
func New(dir string) *Pipeline {
	return &Pipeline{
		fsys: os.DirFS(dir),
		reg:  rules.NewRegistry(),
	}
}

// Run executes the pipeline and returns the generated Dockerfile
func (p *Pipeline) Run() (string, error) {
	// Detect project type
	rule, ok := p.reg.Detect(p.fsys)
	if !ok {
		return "", detect.ErrUnknownStack
	}

	// Extract facts
	f, err := facts.GetFactsFromRule(p.fsys, rule)
	if err != nil {
		return "", err
	}

	// Generate Dockerfile
	return generateDockerfile(f)
}

func generateDockerfile(f types.Facts) (string, error) {
	var sb strings.Builder

	// Base image
	sb.WriteString(fmt.Sprintf("FROM %s\n\n", f.BaseImage))

	// Working directory
	sb.WriteString("WORKDIR /app\n\n")

	// Copy source code
	sb.WriteString("COPY . .\n\n")

	// Environment variables
	if len(f.Env) > 0 {
		for key, value := range f.Env {
			sb.WriteString(fmt.Sprintf("ENV %s=%s\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Build command
	if f.BuildCmd != "" {
		sb.WriteString(fmt.Sprintf("RUN %s\n\n", f.BuildCmd))
	}

	// Expose ports
	if len(f.Ports) > 0 {
		for _, port := range f.Ports {
			sb.WriteString(fmt.Sprintf("EXPOSE %d\n", port))
		}
		sb.WriteString("\n")
	}

	// Start command
	if f.StartCmd != "" {
		sb.WriteString(fmt.Sprintf("CMD %s\n", f.StartCmd))
	}

	return sb.String(), nil
}
