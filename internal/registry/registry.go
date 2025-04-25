package registry

import (
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
)

var all []rules.Rule

// Register adds a rule to the registry
func Register(r rules.Rule) {
	all = append(all, r)
}

// All returns all registered rules
func All() []rules.Rule {
	return all
}

func init() {
	Register(&react.Detector{})
}
