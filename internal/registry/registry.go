package registry

import (
	"sync"

	"github.com/doorcloud/door-ai-dockerise/internal/registry/impl"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

var (
	all             []types.Rule
	defaultRegistry types.Registry
	once            sync.Once
)

// Register adds a rule to the registry
func Register(r types.Rule) {
	all = append(all, r)
}

// All returns all registered rules
func All() []types.Rule {
	return all
}

// Default returns the default registry instance
func Default() types.Registry {
	once.Do(func() {
		reg := impl.NewRegistry()
		for _, rule := range all {
			reg.Register(types.NewDetector(rule))
		}
		defaultRegistry = reg
	})
	return defaultRegistry
}
