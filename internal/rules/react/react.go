package react

import (
	"io/fs"
)

// Detector implements types.Rule.
type Detector struct{}

func (Detector) Name() string {
	return "react"
}

func (Detector) Detect(fsys fs.FS) bool {
	detected, _ := ReactDetector{}.Detect(fsys)
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
