package registry

import (
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

var all []types.Rule

// Register adds a rule to the registry
func Register(r types.Rule) {
	all = append(all, r)
}

// All returns all registered rules
func All() []types.Rule {
	return all
}

func init() {
	Register(&react.Detector{})
}
