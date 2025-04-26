package detectors

import (
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/core"
)

// DefaultDetectors returns a slice of default detectors
func DefaultDetectors() []core.Detector {
	return []core.Detector{
		springboot.NewSpringBootDetector(),
		NewReact(),
	}
}
