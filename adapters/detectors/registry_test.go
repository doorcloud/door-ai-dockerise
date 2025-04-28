package detectors

import (
	"testing"
)

func TestRegistryDetectsSpring(t *testing.T) {
	spec, stack, ok := Detect("testdata/spring_positive/petclinic_maven")
	if !ok || stack != "spring-boot" || spec.BuildTool == "" {
		t.Fatalf("spring not detected or spec empty")
	}
}
