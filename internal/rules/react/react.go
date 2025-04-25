package react

import (
	"bytes"
	"os"
	"path/filepath"
)

// Detector implements the rules.Rule interface for React projects
type Detector struct{}

func (d Detector) Name() string {
	return "react"
}

func (d Detector) Detect(dir string) bool {
	// 1. direct hit
	if hasReactPkg(dir) {
		return true
	}
	// 2. look one level below (covers tests like examples/react)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && hasReactPkg(filepath.Join(dir, e.Name())) {
			return true
		}
	}
	return false
}

func hasReactPkg(p string) bool {
	b, err := os.ReadFile(filepath.Join(p, "package.json"))
	return err == nil && bytes.Contains(b, []byte(`"react"`))
}

func (d Detector) Facts(dir string) map[string]any {
	return map[string]any{
		"language":  "JavaScript",
		"framework": "React",
		"build_cmd": "npm ci && npm run build",
		"artifact":  "build",
		"ports":     []int{3000},
	}
}
