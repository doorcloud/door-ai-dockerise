package spring

import (
	"os"
	"path/filepath"
	"strings"
)

// IsSpringBoot checks if the given path contains a Spring Boot project
func IsSpringBoot(path string) bool {
	// Check for build files
	buildFiles := []string{
		"pom.xml",
		"build.gradle",
		"build.gradle.kts",
	}

	// For Maven multi-module projects, also check application/pom.xml
	mavenModules := []string{
		"application/pom.xml",
		"web/pom.xml",
		"app/pom.xml",
	}

	// First check the root build files
	for _, file := range buildFiles {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			content, err := os.ReadFile(filepath.Join(path, file))
			if err != nil {
				continue
			}

			contentStr := string(content)
			if isRootSpringBootProject(contentStr) {
				return true
			}
		}
	}

	// Then check Maven module build files
	for _, file := range mavenModules {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			content, err := os.ReadFile(filepath.Join(path, file))
			if err != nil {
				continue
			}

			contentStr := string(content)
			if isSpringBootModule(contentStr) {
				return true
			}
		}
	}

	return false
}

// isRootSpringBootProject checks if the root project is a Spring Boot project
func isRootSpringBootProject(content string) bool {
	// Must have Spring Boot parent or platform BOM
	hasParent := strings.Contains(content, "<parent>") &&
		(strings.Contains(content, "<groupId>org.springframework.boot</groupId>") ||
			strings.Contains(content, "<groupId>io.spring.platform</groupId>")) &&
		(strings.Contains(content, "<artifactId>spring-boot-starter-parent</artifactId>") ||
			strings.Contains(content, "<artifactId>platform-bom</artifactId>"))

	// Must have Spring Boot plugin
	hasPlugin := strings.Contains(content, "id 'org.springframework.boot'") ||
		strings.Contains(content, "id(\"org.springframework.boot\")") ||
		(strings.Contains(content, "<groupId>org.springframework.boot</groupId>") &&
			strings.Contains(content, "<artifactId>spring-boot-maven-plugin</artifactId>"))

	// Must have Spring Boot starter dependencies
	hasStarter := strings.Contains(content, "<artifactId>spring-boot-starter-web</artifactId>") ||
		strings.Contains(content, "<artifactId>spring-boot-starter-actuator</artifactId>") ||
		strings.Contains(content, "<artifactId>spring-boot-starter-webflux</artifactId>") ||
		strings.Contains(content, "org.springframework.boot:spring-boot-starter-web:") ||
		strings.Contains(content, "org.springframework.boot:spring-boot-starter-actuator:") ||
		strings.Contains(content, "org.springframework.boot:spring-boot-starter-webflux:")

	// For root projects, we need either:
	// 1. Spring Boot parent + starter dependencies, or
	// 2. Spring Boot plugin + starter dependencies
	return (hasParent || hasPlugin) && hasStarter
}

// isSpringBootModule checks if a module is a Spring Boot module
func isSpringBootModule(content string) bool {
	// Must have Spring Boot plugin
	hasPlugin := strings.Contains(content, "id 'org.springframework.boot'") ||
		strings.Contains(content, "id(\"org.springframework.boot\")") ||
		(strings.Contains(content, "<groupId>org.springframework.boot</groupId>") &&
			strings.Contains(content, "<artifactId>spring-boot-maven-plugin</artifactId>"))

	// Must have Spring Boot starter dependencies
	hasStarter := strings.Contains(content, "<artifactId>spring-boot-starter-web</artifactId>") ||
		strings.Contains(content, "<artifactId>spring-boot-starter-actuator</artifactId>") ||
		strings.Contains(content, "<artifactId>spring-boot-starter-webflux</artifactId>") ||
		strings.Contains(content, "org.springframework.boot:spring-boot-starter-web:") ||
		strings.Contains(content, "org.springframework.boot:spring-boot-starter-actuator:") ||
		strings.Contains(content, "org.springframework.boot:spring-boot-starter-webflux:")

	// For modules, we need both plugin and starter dependencies
	return hasPlugin && hasStarter
}
