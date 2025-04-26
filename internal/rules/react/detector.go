package react

import (
	"encoding/json"
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
	err := fs.WalkDir(fsys, ".", func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return nil
		}

		// Stop descending deeper than 3 segments
		if len(strings.Split(path, string(filepath.Separator))) > 3 {
			return fs.SkipDir
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

			if hasPackage(p.Dependencies, "react") || hasPackage(p.DevDependencies, "react") {
				foundReact = true
			}

			if hasBuildTool(p.Dependencies) || hasBuildTool(p.DevDependencies) {
				foundBuildTool = true
			}

			if foundReact && foundBuildTool {
				return fs.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}
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
