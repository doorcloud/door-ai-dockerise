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
        <version>2.7.0</version>
    </parent>
</project>`,
				"src/main/resources/application.properties": "spring.application.name=test",
			},
			wantInfo: core.StackInfo{
				Name:      "springboot",
				BuildTool: "maven",
				Port:      8080,
				DetectedFiles: []string{
					"pom.xml",
					"src/main/resources/application.properties",
				},
			},
			wantFound: true,
		},
		{
			name: "Gradle project with custom port",
			files: map[string]string{
				"build.gradle": `plugins {
    id 'org.springframework.boot' version '2.7.0'
}`,
				"src/main/resources/application.yml": `server:
  port: 9090`,
			},
			wantInfo: core.StackInfo{
				Name:      "springboot",
				BuildTool: "gradle",
				Port:      9090,
				DetectedFiles: []string{
					"build.gradle",
					"src/main/resources/application.yml",
				},
			},
			wantFound: true,
		},
		{
			name: "Gradle Kotlin project",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    id("org.springframework.boot") version "2.7.0"
}`,
			},
			wantInfo: core.StackInfo{
				Name:      "springboot",
				BuildTool: "gradle",
				Port:      8080,
				DetectedFiles: []string{
					"build.gradle.kts",
				},
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
				assert.Equal(t, tt.wantInfo.Port, gotInfo.Port)
				assert.ElementsMatch(t, tt.wantInfo.DetectedFiles, gotInfo.DetectedFiles)
			}
		})
	}
}
