package core

import (
	"fmt"
)

// Spec defines the stack configuration
type Spec struct {
	Language  string            `yaml:"language" json:"language"`
	Framework string            `yaml:"framework" json:"framework"`
	Version   string            `yaml:"version,omitempty" json:"version,omitempty"`
	BuildTool string            `yaml:"buildTool,omitempty" json:"buildTool,omitempty"`
	Params    map[string]string `yaml:"params,omitempty" json:"params,omitempty"`
	Layered   bool              `yaml:"layered,omitempty"`
}

// Validate returns an error if the spec is invalid
func (s *Spec) Validate() error {
	if s.Language == "" {
		return fmt.Errorf("language is required")
	}
	if s.Params == nil {
		s.Params = make(map[string]string)
	}
	return nil
}
