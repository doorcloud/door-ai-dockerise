package node

import (
	"context"
	"os"

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
func (n *NodeDetector) Detect(ctx context.Context, dir string) (core.StackInfo, error) {
	fsys := os.DirFS(dir)
	if n.d.Detect(fsys) {
		return core.StackInfo{
			Name: "node",
			Meta: map[string]string{
				"runtime": "node",
			},
		}, nil
	}
	return core.StackInfo{}, nil
}
