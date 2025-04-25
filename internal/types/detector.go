package types

import (
	"io/fs"
)

// Detector is an interface for detecting project types
type Detector interface {
	// Detect returns true if the detector matches the project
	Detect(fsys fs.FS) (bool, error)
	// Name returns the name of the detector
	Name() string
}

// Registry is an interface for managing detectors
type Registry interface {
	// Register adds a detector to the registry
	Register(detector Detector)
	// GetDetectors returns all registered detectors
	GetDetectors() []Detector
}
