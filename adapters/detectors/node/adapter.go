package node

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/node"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// NodeDetector implements core.Detector for Node.js projects
type NodeDetector struct {
	d node.NodeDetector
}

// NewNodeDetector creates a new NodeDetector
func NewNodeDetector() *NodeDetector {
	return &NodeDetector{d: node.NodeDetector{}}
}

// Detect implements the core.Detector interface
func (n *NodeDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, bool, error) {
	if n.d.Detect(fsys) {
		return core.StackInfo{
			Name:          "node",
			BuildTool:     "npm",
			DetectedFiles: []string{"package.json"},
		}, true, nil
	}
	return core.StackInfo{}, false, nil
}

// Name returns the detector name
func (n *NodeDetector) Name() string {
	return "node"
}
