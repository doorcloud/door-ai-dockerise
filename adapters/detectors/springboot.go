package detectors

import (
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
)

// NewSpringBootDetector creates a new Spring Boot detector
func NewSpringBootDetector() *springboot.SpringBootDetector {
	return springboot.NewSpringBootDetector()
}
