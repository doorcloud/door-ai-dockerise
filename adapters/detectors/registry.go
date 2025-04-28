package detectors

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// Detector is implemented once per stack.
type Detector interface {
	core.Detector
}

var reg []core.Detector

// Register is called from detector init() blocks.
func Register(d core.Detector) { reg = append(reg, d) }

// Registry returns the list of registered detectors.
func Registry() []core.Detector { return reg }

// Detect runs detectors in registration order and returns the first match.
// The returned spec may already have BuildTool etc. filled by the stack extractor.
func Detect(root string) (core.Spec, string, bool) {
	fsys := os.DirFS(root)
	for _, d := range reg {
		info, found, err := d.Detect(context.Background(), fsys, nil)
		if err != nil {
			continue
		}
		if found {
			return core.Spec{
				Framework: info.Name,
				BuildTool: info.BuildTool,
			}, d.Name(), true
		}
	}
	return core.Spec{}, "", false
}

// Explain prints a summary table (optional CLI flag).
func Explain() string {
	var b strings.Builder
	for _, d := range reg {
		fmt.Fprintf(&b, "%-12s │ %s\n", d.Name(), d.Describe())
	}
	return b.String()
}
