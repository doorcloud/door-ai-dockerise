package react

import (
	"os"
	"path/filepath"
)

// Spec represents the detected configuration for a React project
type Spec struct {
	Language   string
	Framework  string
	BuildTool  string
	Port       string
	DistDir    string
	HealthPath string
}

// Detect analyzes the project directory to determine React-specific configuration
func Detect(projectDir string) (*Spec, error) {
	spec := &Spec{
		Language:   "javascript",
		Framework:  "react",
		Port:       "3001",
		DistDir:    "build",
		HealthPath: "/",
	}

	// Detect build tool
	if _, err := os.Stat(filepath.Join(projectDir, "yarn.lock")); err == nil {
		spec.BuildTool = "yarn"
	} else if _, err := os.Stat(filepath.Join(projectDir, "pnpm-lock.yaml")); err == nil {
		spec.BuildTool = "pnpm"
	} else {
		spec.BuildTool = "npm"
	}

	// Check for Vite configuration
	if _, err := os.Stat(filepath.Join(projectDir, "vite.config.js")); err == nil {
		spec.DistDir = "dist"
	}

	return spec, nil
}
