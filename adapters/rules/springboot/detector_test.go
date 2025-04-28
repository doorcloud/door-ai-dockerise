package springboot

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
)

func TestSpringBootDetector_Detect(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		wantInfo  core.StackInfo
		wantFound bool
		wantErr   bool
	}{
		{
			name: "Maven project with default port",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <properties>
        <java.version>17</java.version>
    </properties>
</project>`,
				"src/main/resources/application.properties": "spring.application.name=test",
			},
			wantInfo: core.StackInfo{
				Name:      "spring-boot",
				BuildTool: "maven",
				Version:   "3.2.0",
				Port:      8080,
				DetectedFiles: []string{
					"pom.xml",
					"src/main/resources/application.properties",
				},
				Confidence: 1.0,
			},
			wantFound: true,
		},
		{
			name: "Gradle project with custom port",
			files: map[string]string{
				"build.gradle": `plugins {
    id 'org.springframework.boot' version '3.2.0'
}`,
				"src/main/resources/application.yml": `server:
  port: 9090`,
			},
			wantInfo: core.StackInfo{
				Name:      "spring-boot",
				BuildTool: "gradle",
				Version:   "3.2.0",
				Port:      9090,
				DetectedFiles: []string{
					"build.gradle",
					"src/main/resources/application.yml",
				},
				Confidence: 1.0,
			},
			wantFound: true,
		},
		{
			name: "Gradle Kotlin project",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    id("org.springframework.boot") version "3.2.0"
}`,
			},
			wantInfo: core.StackInfo{
				Name:      "spring-boot",
				BuildTool: "gradle",
				Version:   "3.2.0",
				Port:      8080,
				DetectedFiles: []string{
					"build.gradle.kts",
				},
				Confidence: 1.0,
			},
			wantFound: true,
		},
		{
			name: "Not a Spring Boot project",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.apache.maven</groupId>
        <artifactId>maven-parent</artifactId>
        <version>1.0</version>
    </parent>
</project>`,
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test filesystem
			fsys := fstest.MapFS{}
			for name, content := range tt.files {
				fsys[name] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			d := NewSpringBootDetector()
			gotInfo, gotFound, err := d.Detect(context.Background(), fsys, nil)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantFound, gotFound)
			if tt.wantFound {
				assert.Equal(t, tt.wantInfo.Name, gotInfo.Name)
				assert.Equal(t, tt.wantInfo.BuildTool, gotInfo.BuildTool)
				assert.Equal(t, tt.wantInfo.Version, gotInfo.Version)
				assert.Equal(t, tt.wantInfo.Port, gotInfo.Port)
				assert.Equal(t, tt.wantInfo.Confidence, gotInfo.Confidence)
				assert.ElementsMatch(t, tt.wantInfo.DetectedFiles, gotInfo.DetectedFiles)
			}
		})
	}
}
