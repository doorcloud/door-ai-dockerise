package impl

import (
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

// Registry implements the types.Registry interface
type Registry struct {
	detectors []types.Detector
}

// NewRegistry creates a new empty registry
func NewRegistry() *Registry {
	return &Registry{
		detectors: make([]types.Detector, 0),
	}
}

// Register adds a detector to the registry
func (r *Registry) Register(detector types.Detector) {
	r.detectors = append(r.detectors, detector)
}

// GetDetectors returns all registered detectors
func (r *Registry) GetDetectors() []types.Detector {
	return r.detectors
}

// Detect tries each registered detector in order until one matches
func (r *Registry) Detect(fsys fs.FS) (types.RuleInfo, bool) {
	for _, d := range r.detectors {
		detected, err := d.Detect(fsys)
		if err != nil {
			continue
		}
		if detected {
			return types.RuleInfo{
				Name: d.Name(),
			}, true
		}
	}
	return types.RuleInfo{}, false
}
