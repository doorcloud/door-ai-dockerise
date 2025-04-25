package types

import "io/fs"

// Rule defines the interface for technology stack detection rules
type Rule interface {
	// Name returns the unique identifier for this rule
	Name() string
	// Detect checks if the given filesystem contains a project matching this rule
	Detect(fsys fs.FS) bool
}
