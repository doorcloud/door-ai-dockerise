package react

import (
	"bytes"
	"io/fs"
)

// Detector implements types.Rule.
type Detector struct{}

func (Detector) Name() string {
	return "react"
}

func (Detector) Detect(fsys fs.FS) bool {
	found := false
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if found || err != nil {
			return fs.SkipDir
		}

		if d.IsDir() && d.Name() == "node_modules" {
			return fs.SkipDir
		}

		if d.Name() == "package.json" && hasReactPkg(fsys, path) {
			found = true
			return fs.SkipDir
		}
		return nil
	})
	return found
}

func hasReactPkg(fsys fs.FS, pkgJSONPath string) bool {
	b, err := fs.ReadFile(fsys, pkgJSONPath)
	if err != nil {
		return false
	}
	// Check for different quote styles
	return bytes.Contains(b, []byte(`"react"`)) || bytes.Contains(b, []byte(`'react'`)) || bytes.Contains(b, []byte(`react:`))
}

func (Detector) Facts(fsys fs.FS) map[string]any {
	return map[string]any{
		"language":  "JavaScript",
		"framework": "React",
		"build_cmd": "npm ci && npm run build",
		"start_cmd": "serve -s build",
		"artifact":  "build",
		"ports":     []int{80},
	}
}
