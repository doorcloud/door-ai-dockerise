package llm

import (
	"bytes"
	"context"
	"html/template"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/config"
	"github.com/pkg/errors"
)

// Generator generates Dockerfiles using LLM
type Generator struct {
	templateDir string
	cfg         *config.Config
}

// New creates a new LLM generator
func New(cfg *config.Config, templateDir string) generate.Generator {
	return &Generator{
		templateDir: templateDir,
		cfg:         cfg,
	}
}

// Generate creates a Dockerfile using the provided facts
func (g *Generator) Generate(ctx context.Context, facts core.Facts) (string, error) {
	// Convert facts to map for template
	factsMap := map[string]interface{}{
		"Framework": facts.StackType,
		"BuildTool": facts.BuildTool,
		"Port":      facts.Port,
	}

	// Determine template based on framework
	templateFile := filepath.Join(g.templateDir, facts.StackType+"_docker_prompt.txt")
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse template %s", templateFile)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, factsMap); err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}

	return buf.String(), nil
}
