package springboot

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/rules"
)

// setupTestDir creates a temporary directory with the given files and mvnw script
func setupTestDir(t *testing.T, files map[string]string) string {
	tmpDir := t.TempDir()

	// Create mvnw script
	mvnwPath := filepath.Join(tmpDir, "mvnw")
	mvnwContent := `#!/bin/sh
# Minimal Maven wrapper for testing
if [ "$1" = "clean" ] && [ "$2" = "package" ] && [ "$3" = "-DskipTests" ]; then
    mkdir -p target
    touch target/app.jar
    echo "Maven build successful"
    exit 0
else
    echo "Unsupported Maven command"
    exit 1
fi`
	if err := os.WriteFile(mvnwPath, []byte(mvnwContent), 0755); err != nil {
		t.Fatalf("Failed to create mvnw: %v", err)
	}

	// Create test files
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	return tmpDir
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantFact rules.Facts
		want     bool
	}{
		{
			name: "Maven project with Spring Boot starter",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>demo</artifactId>
    <version>0.0.1-SNAPSHOT</version>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
            <version>3.2.0</version>
        </dependency>
    </dependencies>
</project>`,
			},
			wantFact: rules.Facts{
				Language:     "java",
				Framework:    "spring-boot",
				BuildTool:    "maven",
				BuildCmd:     "mvn clean package -DskipTests",
				BuildDir:     ".",
				StartCmd:     "java -jar app.jar",
				Artifact:     "target/*.jar",
				Ports:        []int{8080},
				Health:       "/actuator/health",
				BaseHint:     "eclipse-temurin:17-jdk",
				MavenVersion: "3.9.6",
				Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
			want: true,
		},
		{
			name: "maven_with_spring_boot_starter",
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
			wantFact: rules.Facts{
				Language:     "java",
				Framework:    "spring-boot",
				BuildTool:    "maven",
				BuildCmd:     "mvn clean package -DskipTests",
				BuildDir:     ".",
				StartCmd:     "java -jar app.jar",
				Artifact:     "target/*.jar",
				Ports:        []int{8080},
				Health:       "/actuator/health",
				BaseHint:     "eclipse-temurin:17-jdk",
				MavenVersion: "3.9.6",
				Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
			want: true,
		},
		{
			name: "gradle_with_spring_boot_plugin",
			files: map[string]string{
				"build.gradle": `plugins {
    id 'org.springframework.boot' version '2.7.0'
}`,
			},
			wantFact: rules.Facts{
				Language:     "java",
				Framework:    "spring-boot",
				BuildTool:    "gradle",
				BuildCmd:     "./gradlew build -x test",
				BuildDir:     ".",
				StartCmd:     "java -jar app.jar",
				Artifact:     "build/libs/*.jar",
				Ports:        []int{8080},
				Health:       "/actuator/health",
				BaseHint:     "eclipse-temurin:17-jdk",
				MavenVersion: "3.9.6",
				Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
			want: true,
		},
		{
			name: "kotlin_dsl_with_spring_boot",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    id("org.springframework.boot") version "2.7.0"
}`,
			},
			wantFact: rules.Facts{
				Language:     "java",
				Framework:    "spring-boot",
				BuildTool:    "gradle",
				BuildCmd:     "./gradlew build -x test",
				BuildDir:     ".",
				StartCmd:     "java -jar app.jar",
				Artifact:     "build/libs/*.jar",
				Ports:        []int{8080},
				Health:       "/actuator/health",
				BaseHint:     "eclipse-temurin:17-jdk",
				MavenVersion: "3.9.6",
				Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
			want: true,
		},
		{
			name: "application_class_with_annotation",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?><project></project>`,
				"src/main/java/com/example/DemoApplication.java": `
package com.example;
import org.springframework.boot.autoconfigure.SpringBootApplication;
@SpringBootApplication
public class DemoApplication {
}`,
			},
			wantFact: rules.Facts{
				Language:     "java",
				Framework:    "spring-boot",
				BuildTool:    "maven",
				BuildCmd:     "mvn clean package -DskipTests",
				BuildDir:     ".",
				StartCmd:     "java -jar app.jar",
				Artifact:     "target/*.jar",
				Ports:        []int{8080},
				Health:       "/actuator/health",
				BaseHint:     "eclipse-temurin:17-jdk",
				MavenVersion: "3.9.6",
				Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
			want: true,
		},
		{
			name: "multi_module_with_spring_boot",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <modules>
        <module>module1</module>
        <module>module2</module>
    </modules>
</project>`,
				"module1/pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`,
			},
			wantFact: rules.Facts{
				Language:     "java",
				Framework:    "spring-boot",
				BuildTool:    "maven",
				BuildCmd:     "mvn clean package -DskipTests",
				BuildDir:     ".",
				StartCmd:     "java -jar app.jar",
				Artifact:     "target/*.jar",
				Ports:        []int{8080},
				Health:       "/actuator/health",
				BaseHint:     "eclipse-temurin:17-jdk",
				MavenVersion: "3.9.6",
				Env:          map[string]string{"SPRING_PROFILES_ACTIVE": "prod"},
			},
			want: true,
		},
		{
			name: "not_spring_boot",
			files: map[string]string{
				"pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
        </dependency>
    </dependencies>
</project>`,
			},
			wantFact: rules.Facts{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupTestDir(t, tt.files)
			rule := NewRule(slog.Default(), nil, &rules.RuleConfig{})
			if got := rule.Detect(tmpDir); got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}
