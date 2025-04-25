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
	pkg := filepath.Join(dir, "package.json")
	b, err := os.ReadFile(pkg)
	if err != nil {
		return false
	}
	// very small and safe: just look for "react" in dependencies or devDependencies
	return bytes.Contains(b, []byte(`"react"`))
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
