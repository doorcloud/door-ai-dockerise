package springboot

import (
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/registry"
)

func TestSpringRuleRegistered(t *testing.T) {
	if len(registry.All()) == 0 {
		t.Fatal("expected at least one rule registered")
	}
}
