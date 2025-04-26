package react

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// ReactDetector implements types.Rule.
type ReactDetector struct{}

func (ReactDetector) Name() string {
	return "react"
}

// Detect returns true if the project is a React application
func (d *ReactDetector) Detect(fsys fs.FS) bool {
	var foundReact, foundBuildTool bool

	// Walk through the project directory
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		// Skip errors and continue walking
		if err != nil {
			return nil
		}

		// Skip directories
		if d.IsDir() {
			// Skip node_modules
			if d.Name() == "node_modules" {
				return fs.SkipDir
			}
			return nil
		}

		// Check for package.json
		if d.Name() == "package.json" {
			fmt.Printf("Found package.json at %s\n", path)
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				return nil
			}

			var pkg struct {
				Dependencies    map[string]string `json:"dependencies"`
				DevDependencies map[string]string `json:"devDependencies"`
			}
			if err := json.Unmarshal(data, &pkg); err != nil {
				return nil
			}

			fmt.Printf("Dependencies: %v\n", pkg.Dependencies)
			fmt.Printf("DevDependencies: %v\n", pkg.DevDependencies)

			// Check for React dependency
			if _, ok := pkg.Dependencies["react"]; ok {
				fmt.Println("Found react dependency")
				foundReact = true
			}

			// Check for build tool (npm)
			if _, ok := pkg.Dependencies["react-scripts"]; ok {
				fmt.Println("Found build tool")
				foundBuildTool = true
			} else if _, ok := pkg.Dependencies["vite"]; ok {
				fmt.Println("Found build tool")
				foundBuildTool = true
			}

			// If we found both, we can stop walking
			if foundReact && foundBuildTool {
				fmt.Println("Found both react and build tool, stopping search")
				return fs.SkipAll
			}
		}

		return nil
	})

	// Ignore any errors from WalkDir
	if err != nil {
		return false
	}

	fmt.Printf("Detection result - foundReact: %v, foundBuildTool: %v\n", foundReact, foundBuildTool)
	return foundReact && foundBuildTool
}

func hasPackage(m map[string]any, pkg string) bool {
	_, ok := m[pkg]
	return ok
}

func hasBuildTool(m map[string]any) bool {
	for k := range m {
		if k == "react-scripts" || k == "vite" || k == "next" {
			return true
		}
	}
	return false
}

func (d *ReactDetector) Facts(fsys fs.FS) map[string]any {
	return map[string]any{
		"language":  "javascript",
		"framework": "react",
		"build_cmd": "npm ci && npm run build",
		"artifact":  "build",
		"ports":     []int{3000},
		"base_hint": "node:18-alpine",
	}
}
