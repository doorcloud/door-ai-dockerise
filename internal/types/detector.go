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
	// Detect tries each registered detector in order until one matches
	Detect(fsys fs.FS) (RuleInfo, bool)
}

// RuleInfo represents information about a detected technology stack
type RuleInfo struct {
	Name string // e.g. "spring-boot"
	Tool string // e.g. "maven"
}
