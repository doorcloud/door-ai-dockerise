package dockerverify

import (
	"context"
	"io/fs"
	"strings"
)

func Verify(ctx context.Context, fsys fs.FS, dockerfile string) error {
	// Strip markdown fences if present
	dockerfile = strings.TrimSpace(
		strings.TrimPrefix(
			strings.TrimSuffix(dockerfile, "```"),
			"```dockerfile"))

	// Rest of the function remains the same...
	return nil
}
