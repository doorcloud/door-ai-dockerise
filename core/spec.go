package core

// Spec represents the input specification for Dockerfile generation
type Spec struct {
	Stack   string            `yaml:"stack"`   // e.g. "react", "spring-boot"
	Version string            `yaml:"version"` // optional
	Facts   map[string]string `yaml:"facts"`   // free-form key-value
}
