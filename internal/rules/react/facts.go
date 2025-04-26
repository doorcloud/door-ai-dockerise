package react

import (
	"encoding/json"
	"io/fs"
)

// FactsDetector extracts the static bits we need for the LLM.
type FactsDetector struct{}

func (FactsDetector) Name() string {
	return "react"
}

func (FactsDetector) Detect(fsys fs.FS) bool {
	return (&ReactDetector{}).Detect(fsys)
}

func (FactsDetector) Facts(fsys fs.FS) map[string]any {
	data, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return nil
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	buildCmd := "npm run build"
	if _, ok := pkg.Scripts["build"]; !ok {
		buildCmd = "npm install && npm run build"
	}

	return map[string]any{
		"language":  "javascript",
		"framework": "react",
		"build_cmd": buildCmd,
		"artifact":  "build",
		"ports":     "3000",
		"base_hint": "node:18-alpine",
	}
}
