package rule

import (
	"sync"
)

// Registry manages the collection of available rules.
type Registry struct {
	rules map[string]Rule
	mu    sync.RWMutex
}

// NewRegistry creates a new rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]Rule),
	}
}

// Register adds a rule to the registry.
func (r *Registry) Register(name string, rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[name] = rule
}

// Get returns a rule by name.
func (r *Registry) Get(name string) (Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[name]
	return rule, ok
}

// List returns all registered rules.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.rules))
	for name := range r.rules {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry is the global registry instance.
var DefaultRegistry = NewRegistry()

// RegisterDefault adds a rule to the default registry.
func RegisterDefault(name string, rule Rule) {
	DefaultRegistry.Register(name, rule)
}

// GetDefault returns a rule from the default registry.
func GetDefault(name string) (Rule, bool) {
	return DefaultRegistry.Get(name)
}

// ListDefault returns all rules from the default registry.
func ListDefault() []string {
	return DefaultRegistry.List()
}
