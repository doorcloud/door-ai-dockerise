package react

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Detector implements the rules.Rule interface for React projects
type Detector struct{}

func (d *Detector) Name() string {
	return "react"
}

func (d *Detector) Detect(repo string) bool {
	// Check for package.json
	pkgPath := filepath.Join(repo, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}

	// Check for React dependency
	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	if _, hasReact := pkg.Dependencies["react"]; !hasReact {
		return false
	}

	// Check for src/index.js or src/index.tsx
	return fileExists(filepath.Join(repo, "src", "index.js")) ||
		fileExists(filepath.Join(repo, "src", "index.tsx"))
}

func (d *Detector) Facts(repo string) map[string]any {
	return map[string]any{
		"language":  "JavaScript",
		"framework": "React",
		"build_cmd": "npm ci && npm run build",
		"artifact":  "build",
		"ports":     []int{3000},
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
