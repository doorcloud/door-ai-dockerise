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
func (n *NodeDetector) Detect(ctx context.Context, fsys fs.FS) (core.StackInfo, error) {
	if n.d.Detect(fsys) {
		return core.StackInfo{
			Name:      "node",
			BuildTool: "npm",
		}, nil
	}
	return core.StackInfo{}, nil
}
