package springboot

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/rules"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected bool
	}{
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
			expected: true,
		},
		{
			name: "gradle_with_spring_boot_plugin",
			files: map[string]string{
				"build.gradle": `plugins {
    id 'org.springframework.boot' version '2.7.0'
}`,
			},
			expected: true,
		},
		{
			name: "kotlin_dsl_with_spring_boot",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    id("org.springframework.boot") version "2.7.0"
}`,
			},
			expected: true,
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
			expected: true,
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
			expected: true,
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
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create test files
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			}

			// Create rule and test detection
			rule := NewRule(slog.Default(), nil, &rules.RuleConfig{})
			if got := rule.Detect(tmpDir); got != tt.expected {
				t.Errorf("Detect() = %v, want %v", got, tt.expected)
			}
		})
	}
}
