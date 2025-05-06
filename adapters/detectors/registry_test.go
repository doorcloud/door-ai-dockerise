package detectors_test

import (
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
)

func TestRegistryUniqueNames(t *testing.T) {
	seen := map[string]struct{}{}
	for _, d := range detectors.List() {
		if _, ok := seen[d.Name()]; ok {
			t.Fatalf("duplicate detector name %q", d.Name())
		}
		seen[d.Name()] = struct{}{}
	}
}
