package spring

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestSpringDetector(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "Maven Spring Boot project",
			setup: func() string {
				dir := t.TempDir()
				pom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
</project>`
				if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(pom), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			expected: true,
		},
		{
			name: "Gradle Spring Boot project",
			setup: func() string {
				dir := t.TempDir()
				gradle := `plugins {
    id("org.springframework.boot") version "3.2.0"
}`
				if err := os.WriteFile(filepath.Join(dir, "build.gradle"), []byte(gradle), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			expected: true,
		},
		{
			name: "Non-Spring project",
			setup: func() string {
				dir := t.TempDir()
				pom := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
</project>`
				if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(pom), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup()
			fsys := os.DirFS(dir)
			detector := &springDetector{}
			logSink := &core.StringLogSink{}

			info, found, err := detector.Detect(context.Background(), fsys, logSink)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if found != tt.expected {
				t.Errorf("Detect() found = %v, want %v", found, tt.expected)
			}

			if found {
				if info.Name != "spring-boot" {
					t.Errorf("Detect() info.Name = %v, want spring-boot", info.Name)
				}
				if info.BuildTool != "maven" && info.BuildTool != "gradle" {
					t.Errorf("Detect() info.BuildTool = %v, want maven or gradle", info.BuildTool)
				}
			}
		})
	}
}
