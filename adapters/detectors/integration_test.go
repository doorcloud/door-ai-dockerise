package detectors

import (
	"context"
	"os"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestIntegration(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "spring boot maven",
			path:     "testdata/spring/maven-single",
			expected: "spring-boot",
		},
		{
			name:     "spring boot gradle",
			path:     "testdata/spring/gradle-groovy",
			expected: "spring-boot",
		},
		{
			name:     "deep nested kts",
			path:     "testdata/deep_nested_kts",
			expected: "spring-boot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := os.DirFS(tt.path)
			info, found, err := core.Detect(context.Background(), fsys, nil)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if !found {
				t.Fatalf("Detect() found = false, want true")
			}
			if info.Name != tt.expected {
				t.Fatalf("Detect() name = %v, want %v", info.Name, tt.expected)
			}
		})
	}
}
