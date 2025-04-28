package springboot

import (
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
)

func init() {
	detectors.Register(NewSpringBootDetector())
}
