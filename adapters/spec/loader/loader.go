package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doorcloud/door-ai-dockerise/core"
	"gopkg.in/yaml.v3"
)

// Load reads a spec file and returns a Spec
func Load(path string) (*core.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}

	var spec core.Spec
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("validate spec: %w", err)
	}

	return &spec, nil
}
