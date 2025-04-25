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
	// Read package.json
	data, err := os.ReadFile(filepath.Join(repo, "package.json"))
	if err != nil {
		return false
	}

	// Parse package.json
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	// Check for React and build tool
	hasReact := false
	hasVite := false
	hasCRA := false

	// Check dependencies
	for dep := range pkg.Dependencies {
		if dep == "react" {
			hasReact = true
		}
		if dep == "vite" {
			hasVite = true
		}
		if dep == "react-scripts" {
			hasCRA = true
		}
	}

	// Check devDependencies
	for dep := range pkg.DevDependencies {
		if dep == "vite" {
			hasVite = true
		}
		if dep == "react-scripts" {
			hasCRA = true
		}
	}

	// Must have React and either Vite or CRA
	return hasReact && (hasVite || hasCRA)
}

func (d *Detector) Facts(repo string) map[string]any {
	// Check if it's a Vite project
	isVite := false
	if data, err := os.ReadFile(filepath.Join(repo, "package.json")); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			_, hasVite := pkg.Dependencies["vite"]
			_, hasViteDev := pkg.DevDependencies["vite"]
			isVite = hasVite || hasViteDev
		}
	}

	// Set facts based on build tool
	if isVite {
		return map[string]any{
			"language":  "javascript",
			"framework": "react",
			"buildTool": "npm",
			"buildCmd":  "npm run build",
			"buildDir":  "dist",
			"startCmd":  "npm start",
			"artifact":  "dist",
			"ports":     []int{5173},
			"health":    "/",
			"baseImage": "node:18-alpine",
			"env":       map[string]string{},
		}
	}

	// Default to CRA
	return map[string]any{
		"language":  "javascript",
		"framework": "react",
		"buildTool": "npm",
		"buildCmd":  "npm run build",
		"buildDir":  "build",
		"startCmd":  "npm start",
		"artifact":  "build",
		"ports":     []int{3000},
		"health":    "/",
		"baseImage": "node:18-alpine",
		"env":       map[string]string{},
	}
}
