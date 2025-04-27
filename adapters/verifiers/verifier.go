package verifiers

// Verifier defines the interface for verifying Dockerfile generation
type Verifier interface {
	// Verify checks if the generated Dockerfile is valid
	Verify(dockerfilePath string) error
}
