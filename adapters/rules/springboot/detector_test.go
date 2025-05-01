package springboot

import (
	"context"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestSpringBootDetector_Detect(t *testing.T) {
	tests := []struct {
		name   string
		fsys   fstest.MapFS
		want   core.StackInfo
		wantOk bool
	}{
		{
			name: "Maven project with Spring Boot parent",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
  <parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>2.7.0</version>
  </parent>
</project>`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "maven",
				DetectedFiles: []string{"pom.xml"},
			},
			wantOk: true,
		},
		{
			name: "Maven project with Spring Boot dependency",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-web</artifactId>
      <version>2.7.0</version>
    </dependency>
  </dependencies>
</project>`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "maven",
				DetectedFiles: []string{"pom.xml"},
			},
			wantOk: true,
		},
		{
			name: "Gradle project with Spring Boot plugin",
			fsys: fstest.MapFS{
				"build.gradle": &fstest.MapFile{
					Data: []byte(`plugins {
  id 'org.springframework.boot' version '2.7.0'
}`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle"},
			},
			wantOk: true,
		},
		{
			name: "Gradle project with Spring Boot dependency",
			fsys: fstest.MapFS{
				"build.gradle": &fstest.MapFile{
					Data: []byte(`dependencies {
  implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
}`),
				},
			},
			want: core.StackInfo{
				Name:          "spring-boot",
				Version:       "2.7.0",
				Confidence:    1.0,
				BuildTool:     "gradle",
				DetectedFiles: []string{"build.gradle"},
			},
			wantOk: true,
		},
		{
			name: "Not a Spring Boot project",
			fsys: fstest.MapFS{
				"pom.xml": &fstest.MapFile{
					Data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>example</artifactId>
      <version>1.0.0</version>
    </dependency>
  </dependencies>
</project>`),
				},
			},
			want:   core.StackInfo{},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewSpringBootDetector()
			got, ok, err := detector.Detect(context.Background(), tt.fsys, nil)
			if err != nil {
				t.Errorf("SpringBootDetector.Detect() error = %v", err)
				return
			}
			if ok != tt.wantOk {
				t.Errorf("SpringBootDetector.Detect() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SpringBootDetector.Detect() = %v, want %v", got, tt.want)
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
	info, found, err := detector.Detect(context.Background(), fs, &core.NullLogSink{})
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if !found {
		t.Fatal("Expected to find Spring Boot project")
	}

	// Verify the result
	if info.Name != "spring-boot" {
		t.Errorf("Expected stack name 'spring-boot', got '%s'", info.Name)
	}
	if info.Version != "3.2.0" {
		t.Errorf("Expected version '3.2.0', got '%s'", info.Version)
	}
	if info.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %f", info.Confidence)
	}
}
