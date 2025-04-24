package snippet

import (
	"fmt"
	"strings"
)

// ValidateBaseImage checks that the first FROM line honours baseHint.
// Returns nil if OK, otherwise an error so the caller can fail fast.
func ValidateBaseImage(dockerfile, baseHint string) error {
	if baseHint == "" {
		return nil // nothing to verify
	}
	first := strings.TrimSpace(strings.SplitN(dockerfile, "\n", 2)[0])
	if !strings.Contains(first, baseHint) {
		return fmt.Errorf("LLM-selected base image %q does not match hint %q", first, baseHint)
	}
	return nil
}
