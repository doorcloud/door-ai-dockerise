package react

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
)

// Detector implements types.Rule.
type Detector struct{}

func (Detector) Name() string {
	return "react"
}

func (Detector) Detect(fsys fs.FS) bool {
	detected, _ := detectReact(fsys)
	return detected
}

func (Detector) Facts(fsys fs.FS) map[string]any {
	return map[string]any{
		"language":   "javascript",
		"framework":  "react",
		"build_tool": "npm",
		"build_cmd":  "npm ci && npm run build",
		"build_dir":  ".",
		"ports":      []int{3000},
		"base_hint":  "node:18-alpine",
	}
}

// ReactDetector implements types.Detector.
type ReactDetector struct{}

func (ReactDetector) Name() string {
	return "react"
}

func (ReactDetector) Detect(fsys fs.FS) (bool, error) {
	return detectReact(fsys)
}

// detectReact is the shared implementation for both detectors
func detectReact(fsys fs.FS) (bool, error) {
	var found bool
	err := fs.WalkDir(fsys, ".", func(path string, e fs.DirEntry, err error) error {
		if found || err != nil || e.IsDir() {
			return nil
		}

		if strings.EqualFold(filepath.Base(path), "package.json") {
			b, err := fs.ReadFile(fsys, path)
			if err != nil {
				return nil
			}

			var p struct {
				Dependencies    map[string]any `json:"dependencies"`
				DevDependencies map[string]any `json:"devDependencies"`
			}
			if err := json.Unmarshal(b, &p); err != nil {
				return nil
			}

			if hasReact(p.Dependencies) || hasReact(p.DevDependencies) {
				found = true
				return fs.SkipDir
			}
		}

		// Stop descending deeper than 3 segments
		if len(strings.Split(path, string(filepath.Separator))) > 3 {
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		return false, err
	}
	return found, nil
}

func hasReact(m map[string]any) bool {
	for k := range m {
		if k == "react" || k == "react-dom" || k == "next" || k == "react-scripts" {
			return true
		}
	}
	return false
}
