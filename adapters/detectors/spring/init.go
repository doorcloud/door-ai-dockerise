package spring

import (
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
)

func init() {
	detector := NewSpringBootDetectorV2()
	detectors.Register(detector)
}
