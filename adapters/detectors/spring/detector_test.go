package spring

import (
	"context"
	"os"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpringBootDetectorV2(t *testing.T) {
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
				DetectedFiles: []string{"app/build.gradle"},
			},
			want: true,
		},
		{
			name:     "not spring boot",
			project:  "testdata/spring/not-spring",
			wantInfo: core.StackInfo{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a filesystem for the test project
			fsys := os.DirFS(tt.project)

			// Create detector
			detector := NewSpringBootDetectorV2()

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
