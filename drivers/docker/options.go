package docker

// BuildOptions contains options for building a Docker image
type BuildOptions struct {
	// Tags are the tags to apply to the built image
	Tags []string
	// Context is the build context directory
	Context string
	// Dockerfile is the path to the Dockerfile
	Dockerfile string
}
