package registry

import (
	"github.com/doorcloud/door-ai-dockerise/internal/rules/node"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
	"github.com/doorcloud/door-ai-dockerise/internal/rules/springboot"
)

func init() {
	Register(&react.Detector{})
	Register(&springboot.Rule{})
	Register(&node.Detector{})
}
