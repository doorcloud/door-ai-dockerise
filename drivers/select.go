package drivers

import (
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/doorcloud/door-ai-dockerise/drivers/k8sjob"
)

// Select returns the appropriate build driver based on the BUILD_DRIVER environment variable
func Select(driver string) (core.BuildDriver, error) {
	switch driver {
	case "docker", "":
		return docker.NewEngine()
	case "k8s":
		return k8sjob.NewJob()
	default:
		return nil, fmt.Errorf("unsupported build driver: %s", driver)
	}
}

// Default returns the default build driver (Docker engine)
func Default() core.BuildDriver {
	driver, err := Select(os.Getenv("BUILD_DRIVER"))
	if err != nil {
		// Fallback to Docker engine
		driver, _ = Select("docker")
	}
	return driver
}
