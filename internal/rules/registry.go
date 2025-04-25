package rules

import (
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

// Detector defines the interface for project type detection
type Detector interface {
	Detect(fsys fs.FS) (detect.Rule, bool)
}

// Registry holds a list of registered detectors
type Registry struct {
	detectors []Detector
}

// NewRegistry creates a new empty registry
func NewRegistry() *Registry {
	return &Registry{
		detectors: make([]Detector, 0),
	}
}

// Register adds a new detector to the registry
func (r *Registry) Register(d Detector) {
	r.detectors = append(r.detectors, d)
}

// Detect tries each registered detector in order until one matches
func (r *Registry) Detect(fsys fs.FS) (detect.Rule, bool) {
	for _, d := range r.detectors {
		if rule, ok := d.Detect(fsys); ok {
			return rule, true
		}
	}
	return detect.Rule{}, false
}
