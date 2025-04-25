package registry

import (
	"github.com/doorcloud/door-ai-dockerise/internal/rules/react"
)

func init() {
	Register(&react.Detector{})
}
