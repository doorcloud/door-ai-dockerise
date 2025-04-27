package node

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/rules/node"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// NodeDetector implements core.Detector for Node.js projects
type NodeDetector struct {
	d       node.NodeDetector
	logSink core.LogSink
}

// NewNodeDetector creates a new NodeDetector
func NewNodeDetector() *NodeDetector {
	return &NodeDetector{d: node.NodeDetector{}}
}

// Detect implements the core.Detector interface
func (n *NodeDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	n.logSink = logSink

	if n.d.Detect(fsys) {
		info := core.StackInfo{
			Name:          "node",
			BuildTool:     "npm",
			DetectedFiles: []string{"package.json"},
		}
		if logSink != nil {
			logSink.Log("detector=node found=true path=package.json")
		}
		return info, true, nil
	}
	return core.StackInfo{}, false, nil
}

// Name returns the detector name
func (n *NodeDetector) Name() string {
	return "node"
}

// SetLogSink sets the log sink for the detector
func (n *NodeDetector) SetLogSink(logSink core.LogSink) {
	n.logSink = logSink
}
