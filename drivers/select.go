package drivers

import (
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/drivers/k8sjob"
)

// Default returns the default build driver (Docker)
func Default() core.BuildDriver {
	engine, err := docker.NewEngine()
	if err != nil {
		return nil
	}
	return engine
}

// Select returns a build driver based on the driver name
func Select(driver string) core.BuildDriver {
	switch driver {
	case "docker":
		return Default()
	case "k8s":
		return k8sjob.NewJob()
	default:
		return Default()
	}
}
