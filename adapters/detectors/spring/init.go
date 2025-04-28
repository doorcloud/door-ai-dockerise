package spring

import (
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/core"
)

func init() {
	detector := NewSpringBootDetectorV2()
	detectors.Register(detector)
	core.RegisterDetector(detector)
}
