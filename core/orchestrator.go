package core

import (
	"context"
	"io"
)

// Orchestrator coordinates the Dockerfile generation process
type Orchestrator interface {
	// Run executes the complete Dockerfile generation workflow:
	// 1. Detect stack type
	// 2. Gather facts
	// 3. Generate Dockerfile
	// 4. Verify result
	// Logs are streamed to the provided writer
	Run(
		ctx context.Context,
		root string,
		spec *Spec,
		logs io.Writer,
	) (string /*dockerfile*/, error)
}
