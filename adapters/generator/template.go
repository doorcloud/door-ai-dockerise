package generator

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"text/template"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// TemplateGenerator generates Dockerfiles from templates
type TemplateGenerator struct {
	templates map[string]string
	rootDir   string
}

// NewTemplateGenerator creates a new template generator
func NewTemplateGenerator() *TemplateGenerator {
	// Get the absolute path to the project root directory
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	// If we're in a test directory, go up to the project root
	if filepath.Base(filepath.Dir(wd)) == "test" {
		wd = filepath.Dir(filepath.Dir(wd))
	}

	rootDir := filepath.Join(wd, "templates")

	return &TemplateGenerator{
		templates: map[string]string{
			"springboot": "springboot-jar.tmpl",
		},
		rootDir: rootDir,
	}
}

// Generate generates a Dockerfile from a template
func (g *TemplateGenerator) Generate(ctx context.Context, facts core.Facts) (string, error) {
	templatePath, ok := g.templates[facts.StackType]
	if !ok {
		return "", nil
	}

	tmpl, err := template.ParseFiles(filepath.Join(g.rootDir, templatePath))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, facts); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Fix attempts to fix a Dockerfile based on build error
func (g *TemplateGenerator) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	// TODO: Implement fix logic
	return prevDockerfile, nil
}
