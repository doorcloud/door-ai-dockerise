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
			name: "Maven project with Spring Boot parent",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>2.7.0</version>
  </parent>
</project>`,
			},
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "maven",
				DetectedFiles: []string{"pom.xml"},
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name: "Maven project with Spring Boot dependency",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-web</artifactId>
      <version>2.7.0</version>
    </dependency>
  </dependencies>
</project>`,
			},
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "maven",
				DetectedFiles: []string{"pom.xml"},
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name: "Gradle project with Spring Boot plugin",
			files: map[string]string{
				"build.gradle": `plugins {
  id 'org.springframework.boot' version '2.7.0'
}`,
			},
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle"},
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name: "Gradle project with Spring Boot dependency",
			files: map[string]string{
				"build.gradle": `dependencies {
  implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
}`,
			},
			wantInfo: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle"},
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name: "Not a Spring Boot project",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>example</artifactId>
      <version>1.0.0</version>
    </dependency>
  </dependencies>
</project>`,
			},
			wantInfo:  core.StackInfo{},
			wantFound: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			for name, content := range tt.files {
				fsys[name] = &fstest.MapFile{
					Data: []byte(content),
				}
			}

			detector := NewSpringBootDetector()
			gotInfo, gotFound, err := detector.Detect(context.Background(), fsys, &core.NullLogSink{})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantFound, gotFound)
			if tt.wantFound {
				assert.Equal(t, tt.wantInfo, gotInfo)
			}
		})
	}
}

func TestSpringBootDetector_Detect_Basic(t *testing.T) {
	// Create a test filesystem
	fs := fstest.MapFS{
		"pom.xml": &fstest.MapFile{
			Data: []byte(`
<project>
	<parent>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-parent</artifactId>
		<version>3.2.0</version>
	</parent>
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
		</dependency>
	</dependencies>
</project>
`),
		},
	}

	// Create a detector with a null log sink
	detector := &SpringBootDetector{
		logSink: &core.NullLogSink{},
	}

	// Test detection
	info, err := detector.Detect(fs)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Verify the result
	if info.StackName != "spring-boot" {
		t.Errorf("Expected stack name 'spring-boot', got '%s'", info.StackName)
	}
	if info.Version != "3.2.0" {
		t.Errorf("Expected version '3.2.0', got '%s'", info.Version)
	}
	if info.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %f", info.Confidence)
	}
}
