package types

import "io/fs"

// RuleAdapter adapts a Rule to implement the Detector interface
type RuleAdapter struct {
	Rule Rule
}

func (a RuleAdapter) Name() string {
	return a.Rule.Name()
}

func (a RuleAdapter) Detect(fsys fs.FS) (bool, error) {
	return a.Rule.Detect(fsys), nil
}

// NewDetector creates a new Detector from a Rule
func NewDetector(r Rule) Detector {
	return RuleAdapter{Rule: r}
}
