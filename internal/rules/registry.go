package rules

import (
	"errors"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

// ErrUnknownStack is returned when no rule matches the project
var ErrUnknownStack = errors.New("unknown technology stack")

// DetectStack tries to detect the technology stack in the given repository
func DetectStack(fsys fs.FS) (*detect.Rule, error) {
	// Create detector
	detector := &detect.SpringDetector{}

	// Run detection
	rule, ok := detector.Detect(fsys)
	if !ok {
		return nil, ErrUnknownStack
	}

	return &rule, nil
}

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

var all []Rule

// Register is called from each rule's init()
func Register(r Rule) { all = append(all, r) }

// Detect returns the first matching rule, or nil if none match
func Detect(repo string) Rule {
	for _, r := range all {
		if r.Detect(repo) {
			return r
		}
	}
	return nil
}
