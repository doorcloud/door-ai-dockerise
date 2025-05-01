package spring

import (
	"io/fs"
	"os"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
)

func TestDetectSpringBoot(t *testing.T) {
	tests := []struct {
		name     string
		fsys     fs.FS
		want     bool
		wantInfo core.StackInfo
	}{
		{
			name: "Maven single module",
			fsys: os.DirFS("testdata/spring/maven-single"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"pom.xml"},
				Confidence:    1.0,
			},
		},
		{
			name: "Gradle Groovy",
			fsys: os.DirFS("testdata/spring/gradle-groovy"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"build.gradle"},
				Confidence:    1.0,
			},
		},
		{
			name: "Gradle Kotlin",
			fsys: os.DirFS("testdata/spring/gradle-kotlin"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"build.gradle.kts"},
				Confidence:    1.0,
			},
		},
		{
			name: "Maven multi-module",
			fsys: os.DirFS("testdata/spring/maven-multi"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"pom.xml"},
				Confidence:    1.0,
			},
		},
		{
			name: "Gradle multi-module",
			fsys: os.DirFS("testdata/spring/gradle-multi"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"app/build.gradle"},
				Confidence:    1.0,
			},
		},
		{
			name: "Version catalog with alias",
			fsys: os.DirFS("testdata/spring_version_catalog_alias"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"gradle/libs.versions.toml"},
				Confidence:    1.0,
			},
		},
		{
			name: "Settings with alias",
			fsys: os.DirFS("testdata/settings_alias_plugin"),
			want: true,
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Version:       "3.2.0",
				Port:          8080,
				DetectedFiles: []string{"settings.gradle"},
				Confidence:    1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetectorV3()
			info, found := detector.Detect(tt.fsys)
			assert.Equal(t, tt.want, found)
			if found {
				assert.Equal(t, tt.wantInfo, info)
			}
		})
	}
}
