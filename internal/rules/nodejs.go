package rules

import (
	"encoding/json"
	"io/fs"
)

// NodeJS implements the types.Detector interface for Node.js projects
type NodeJS struct{}

func (n *NodeJS) Name() string { return "nodejs" }

// Detect returns true if package.json exists whose "dependencies" or "devDependencies"
// includes express or koa OR if scripts has "start" / "dev".
func (n *NodeJS) Detect(fsys fs.FS) (bool, error) {
	// Check for package.json
	data, err := fs.ReadFile(fsys, "package.json")
	if err != nil {
		return false, nil
	}

	var pkg struct {
		Scripts      map[string]string `json:"scripts"`
		Dependencies map[string]any    `json:"dependencies"`
		DevDeps      map[string]any    `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return false, nil
	}

	// Check for server dependencies
	hasServerDep := func(m map[string]any) bool {
		_, e := m["express"]
		_, k := m["koa"]
		return e || k
	}

	if hasServerDep(pkg.Dependencies) || hasServerDep(pkg.DevDeps) {
		return true, nil
	}

	// Check for start/dev scripts
	for k := range pkg.Scripts {
		if k == "start" || k == "dev" {
			return true, nil
		}
	}

	return false, nil
}

func init() {
	NewRegistry().Register(&NodeJS{})
}
