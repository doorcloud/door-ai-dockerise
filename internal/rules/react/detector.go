package react

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// ReactDetector implements types.Detector.
type ReactDetector struct{}

func (ReactDetector) Name() string {
	return "react"
}

func (ReactDetector) Detect(fsys fs.FS) (bool, error) {
	var foundReact, foundBuildTool bool
	fmt.Println("Starting React detection...")

	err := fs.WalkDir(fsys, ".", func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error walking path %s: %v\n", path, err)
			return nil
		}
		if e.IsDir() {
			return nil
		}

		// Count directory depth (excluding filename)
		dirPath := filepath.Dir(path)
		if dirPath == "." {
			dirPath = ""
		}
		depth := len(strings.Split(dirPath, string(filepath.Separator)))
		fmt.Printf("Checking path: %s (depth: %d)\n", path, depth)

		if depth > 3 {
			fmt.Printf("Skipping path %s - too deep (depth: %d)\n", path, depth)
			return fs.SkipDir
		}

		if strings.EqualFold(filepath.Base(path), "package.json") {
			fmt.Printf("Found package.json at %s\n", path)
			b, err := fs.ReadFile(fsys, path)
			if err != nil {
				fmt.Printf("Error reading package.json at %s: %v\n", path, err)
				return nil
			}

			var p struct {
				Dependencies    map[string]any `json:"dependencies"`
				DevDependencies map[string]any `json:"devDependencies"`
			}
			if err := json.Unmarshal(b, &p); err != nil {
				fmt.Printf("Error parsing package.json at %s: %v\n", path, err)
				return nil
			}

			fmt.Printf("Dependencies: %v\n", p.Dependencies)
			fmt.Printf("DevDependencies: %v\n", p.DevDependencies)

			if hasPackage(p.Dependencies, "react") || hasPackage(p.DevDependencies, "react") {
				fmt.Println("Found react dependency")
				foundReact = true
			}

			if hasBuildTool(p.Dependencies) || hasBuildTool(p.DevDependencies) {
				fmt.Println("Found build tool")
				foundBuildTool = true
			}

			if foundReact && foundBuildTool {
				fmt.Println("Found both react and build tool, stopping search")
				return fs.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error during directory walk: %v\n", err)
		return false, err
	}

	fmt.Printf("Detection result - foundReact: %v, foundBuildTool: %v\n", foundReact, foundBuildTool)
	return foundReact && foundBuildTool, nil
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
