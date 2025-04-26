package node

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/node"
)

// NodeDetector implements core.Detector for Node.js projects
type NodeDetector struct {
	detector node.NodeDetector
}

// NewNodeDetector creates a new NodeDetector
func NewNodeDetector() *NodeDetector {
	return &NodeDetector{
		detector: node.NodeDetector{},
	}
}

// Detect implements the core.Detector interface
func (d *NodeDetector) Detect(ctx context.Context, path string) (core.StackInfo, error) {
	fsys := os.DirFS(path)
	if d.detector.Detect(fsys) {
		facts := d.detector.Facts(fsys)
		return core.StackInfo{
			Name: "node",
			Meta: map[string]string{
				"framework": "node",
				"buildTool": facts["buildTool"].(string),
			},
		}, nil
	}
	return core.StackInfo{}, nil
}
