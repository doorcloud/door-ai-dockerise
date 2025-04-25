package rules

import (
	"context"
	"io/fs"
)

// Rule defines the interface for technology stack rules
type Rule interface {
	// Detect checks if the given filesystem matches this rule
	Detect(ctx context.Context, fsys fs.FS) (bool, error)

	// Snippets extracts relevant code snippets from the filesystem
	Snippets(ctx context.Context, fsys fs.FS) ([]string, error)

	// Facts extracts facts about the technology stack
	Facts(ctx context.Context, fsys fs.FS) (map[string]interface{}, error)

	// Dockerfile generates a Dockerfile for the technology stack
	Dockerfile(ctx context.Context, fsys fs.FS) (string, error)
}

// Detect finds the appropriate rule for the given filesystem
func Detect(ctx context.Context, fsys fs.FS) (Rule, error) {
	// TODO: Implement rule detection logic
	return nil, nil
}
