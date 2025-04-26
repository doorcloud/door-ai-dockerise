package react

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"strings"
)

// Detector implements types.Detector.
type Detector struct{}

func (Detector) Name() string {
	return "react"
}

func (Detector) Detect(fsys fs.FS) (bool, error) {
	// Look for package.json in the root directory
	_, err := fs.Stat(fsys, "package.json")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (Detector) Facts(fsys fs.FS) map[string]any {
	facts := map[string]any{
		"language":  "JavaScript",
		"framework": "React",
		"build_cmd": "npm ci && npm run build",
		"start_cmd": "serve -s build",
		"artifact":  "build",
		"ports":     []int{80},
		"port":      3000, // Default port
	}

	// Read package.json
	data, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return facts
	}

	var pkg struct {
		Scripts struct {
			Start string `json:"start"`
		} `json:"scripts"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return facts
	}

	// Set port based on start script
	if strings.Contains(pkg.Scripts.Start, "vite") {
		facts["port"] = 5173
	} else if strings.Contains(pkg.Scripts.Start, "react-scripts") {
		facts["port"] = 3000
	}

	return facts
}

func hasReactPkg(fsys fs.FS, pkgJSONPath string) bool {
	b, err := fs.ReadFile(fsys, pkgJSONPath)
	if err != nil {
		return false
	}
	// Check for different quote styles
	return bytes.Contains(b, []byte(`"react"`)) || bytes.Contains(b, []byte(`'react'`)) || bytes.Contains(b, []byte(`react:`))
}
