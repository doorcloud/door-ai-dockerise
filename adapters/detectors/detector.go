package detectors

import (
	"context"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/node"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
	"github.com/doorcloud/door-ai-dockerise/core"
)

var registry = []core.Detector{
	spring.NewSpringBootDetectorV2(),
	react.NewReactDetector(),
	node.NewNodeDetector(),
}

// Registry returns the list of available detectors
func Registry() []core.Detector {
	return registry
}

// Detect returns the first stack info that matches, or empty if none.
func Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	for _, detector := range registry {
		info, found, err := detector.Detect(ctx, fsys, logSink)
		if err != nil {
			return core.StackInfo{}, false, err
		}
		if found {
			return info, true, nil
		}
	}
	return core.StackInfo{}, false, nil
}
