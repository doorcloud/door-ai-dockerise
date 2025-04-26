package registry

import (
	"github.com/doorcloud/door-ai-dockerise/internal/rules/node"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/springboot"
)

func init() {
	Register(&react.ReactDetector{})
	Register(&react.FactsDetector{})
	Register(&react.DockerfileGenerator{})
	Register(&springboot.Detector{})
	Register(&node.Detector{})
}
