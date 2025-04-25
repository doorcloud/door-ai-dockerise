package pipeline

import (
	"io/fs"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/facts"
)

// Pipeline represents the Dockerfile generation pipeline
type Pipeline struct {
	fsys fs.FS
}

// New creates a new pipeline for the given directory
func New(dir string) *Pipeline {
	return &Pipeline{
		fsys: os.DirFS(dir),
	}
}

// Run executes the pipeline and returns the generated Dockerfile
func (p *Pipeline) Run() (string, error) {
	// Detect project type
	rule, err := detect.Detect(p.fsys)
	if err != nil {
		return "", err
	}

	// Extract facts
	f, err := facts.GetFactsFromRule(p.fsys, rule)
	if err != nil {
		return "", err
	}

	// Generate Dockerfile
	return generateDockerfile(f)
}

func generateDockerfile(f facts.Facts) (string, error) {
	// TODO: Implement Dockerfile generation
	return "", nil
}
