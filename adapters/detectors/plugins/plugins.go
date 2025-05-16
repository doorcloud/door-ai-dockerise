package plugins

// Import each detector package for its init-side registration.
import (
	_ "github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
	// _ "github.com/doorcloud/door-ai-dockerise/adapters/detectors/node"
	// _ "github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
)
