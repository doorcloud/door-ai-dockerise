package spring_test

import (
	"context"
	"os"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpringBootDetectorV3(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		wantInfo core.StackInfo
		want     bool
	}{
		{
			name:    "maven single module",
			project: "testdata/spring/maven-single",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"pom.xml"},
			},
			want: true,
		},
		{
			name:    "maven multi module",
			project: "testdata/spring/maven-multi",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"app/pom.xml"},
			},
			want: true,
		},
		{
			name:    "gradle groovy",
			project: "testdata/spring/gradle-groovy",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"build.gradle"},
			},
			want: true,
		},
		{
			name:    "gradle kotlin",
			project: "testdata/spring/gradle-kotlin",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"build.gradle.kts"},
			},
			want: true,
		},
		{
			name:    "gradle multi module",
			project: "testdata/spring/gradle-multi",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"app/build.gradle"},
			},
			want: true,
		},
		{
			name:    "deep nested kts",
			project: "testdata/deep_nested_kts",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"sub/sub/build.gradle.kts"},
			},
			want: true,
		},
		{
			name:     "not spring boot",
			project:  "testdata/spring/not-spring",
			wantInfo: core.StackInfo{},
			want:     false,
		},
		{
			name:    "mixed builders",
			project: "testdata/mixed_builders",
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Port:          8080,
				Version:       "3.2.0",
				Confidence:    1.0,
				DetectedFiles: []string{"pom.xml"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a filesystem for the test project
			fsys := os.DirFS(tt.project)

			// Create detector
			detector := spring.NewSpringBootDetectorV3()

			// Run detection
			info, found, err := detector.Detect(context.Background(), fsys, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.want, found)
			if found {
				assert.Equal(t, tt.wantInfo, info)
			}
		})
	}
}

func TestIsSpringBoot(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "deep nested kts",
			path:     "testdata/deep_nested_kts",
			expected: true,
		},
		// ... existing test cases ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := spring.IsSpringBoot(tt.path); got != tt.expected {
				t.Errorf("IsSpringBoot() = %v, want %v", got, tt.expected)
			}
		})
	}
}
