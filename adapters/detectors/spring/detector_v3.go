package spring

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

const (
	maxDepthV3    = 4
	defaultPortV3 = 8080
)

var (
	springGradleRXV3 = regexp.MustCompile(`(?i)org\.springframework\.boot`)
	versionRXV3      = regexp.MustCompile(`<version>(\d+\.\d+\.\d+)(?:-[^<]+)?</version>|version\s+['"]([\d.]+)(?:-[^'"]+)?['"]`)
)

// SpringBootDetectorV3 implements core.Detector for Spring Boot projects
type SpringBootDetectorV3 struct {
	logSink core.LogSink
}

// NewSpringBootDetectorV3 creates a new Spring Boot detector
func NewSpringBootDetectorV3() *SpringBootDetectorV3 {
	return &SpringBootDetectorV3{}
}

// Detect implements the Detector interface
func (d *SpringBootDetectorV3) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink

	// First check for Maven projects
	if info, found, err := d.detectMaven(fsys); err != nil {
		return core.StackInfo{}, false, err
	} else if found {
		return info, true, nil
	}

	// Then check for Gradle projects
	if info, found, err := d.detectGradle(fsys); err != nil {
		return core.StackInfo{}, false, err
	} else if found {
		return info, true, nil
	}

	return core.StackInfo{}, false, nil
}

// detectMaven checks for Maven projects (single and multi-module)
func (d *SpringBootDetectorV3) detectMaven(fsys fs.FS) (core.StackInfo, bool, error) {
	var pomPaths []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			depth := strings.Count(path, string(filepath.Separator))
			if depth > maxDepthV3 {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Base(path) == "pom.xml" {
			pomPaths = append(pomPaths, path)
		}
		return nil
	})
	if err != nil {
		return core.StackInfo{}, false, err
	}

	// Check all pom.xml files
	for _, pomPath := range pomPaths {
		if info, found := d.isSpringBootMavenModule(fsys, pomPath); found {
			info.DetectedFiles = []string{pomPath}
			d.log("detector=spring-boot found=true buildtool=maven confidence=%f", info.Confidence)
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// detectGradle checks for Gradle projects (Groovy, Kotlin, and multi-module)
func (d *SpringBootDetectorV3) detectGradle(fsys fs.FS) (core.StackInfo, bool, error) {
	// First check if there's a pom.xml to prefer Maven
	if _, err := fs.ReadFile(fsys, "pom.xml"); err == nil {
		return core.StackInfo{}, false, nil
	}

	var gradlePaths []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			depth := strings.Count(path, string(filepath.Separator))
			if depth > maxDepthV3 {
				return fs.SkipDir
			}
			return nil
		}
		name := filepath.Base(path)
		if strings.HasPrefix(name, "build.gradle") {
			gradlePaths = append(gradlePaths, path)
		}
		return nil
	})
	if err != nil {
		return core.StackInfo{}, false, err
	}

	// Check all Gradle files
	for _, gradlePath := range gradlePaths {
		if info, found := d.isSpringBootGradleModule(fsys, gradlePath); found {
			info.DetectedFiles = []string{gradlePath}
			d.log("detector=spring-boot found=true buildtool=gradle confidence=%f", info.Confidence)
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// isSpringBootMavenModule checks if the pom.xml is a Spring Boot module
func (d *SpringBootDetectorV3) isSpringBootMavenModule(fsys fs.FS, pomPath string) (core.StackInfo, bool) {
	content, err := fs.ReadFile(fsys, pomPath)
	if err != nil {
		return core.StackInfo{}, false
	}

	contentStr := string(content)
	signals := 0

	// Check for Spring Boot parent or platform BOM
	hasParent := strings.Contains(contentStr, "<parent>") &&
		(strings.Contains(contentStr, "<groupId>org.springframework.boot</groupId>") ||
			strings.Contains(contentStr, "<groupId>io.spring.platform</groupId>")) &&
		(strings.Contains(contentStr, "<artifactId>spring-boot-starter-parent</artifactId>") ||
			strings.Contains(contentStr, "<artifactId>platform-bom</artifactId>"))
	if hasParent {
		signals++
	}

	// Check for Spring Boot plugin
	hasPlugin := strings.Contains(contentStr, "<groupId>org.springframework.boot</groupId>") &&
		strings.Contains(contentStr, "<artifactId>spring-boot-maven-plugin</artifactId>")
	if hasPlugin {
		signals++
	}

	// Check for Spring Boot starter dependencies
	hasStarter := strings.Contains(contentStr, "<artifactId>spring-boot-starter-web</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-actuator</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-webflux</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-data-jpa</artifactId>")
	if hasStarter {
		signals++
	}

	// Check for Spring Boot annotations in Java files
	hasAnnotations := d.checkForSpringBootAnnotations(fsys, filepath.Dir(pomPath))
	if hasAnnotations {
		signals++
	}

	// Check for application properties/yml
	hasConfig := d.checkForSpringBootConfig(fsys, filepath.Dir(pomPath))
	if hasConfig {
		signals++
	}

	// For Maven modules, we need either:
	// 1. Spring Boot parent + starter dependencies, or
	// 2. Spring Boot plugin + starter dependencies
	if !((hasParent || hasPlugin) && hasStarter) {
		return core.StackInfo{}, false
	}

	// Calculate confidence based on signals
	confidence := 0.5
	if signals >= 3 {
		confidence = 1.0
	} else if signals == 2 {
		confidence = 0.8
	}

	// Extract version if available
	version := ""
	if hasParent {
		version = d.extractVersionFromMaven(contentStr)
	}

	return core.StackInfo{
		Name:       "spring-boot",
		BuildTool:  "maven",
		Port:       defaultPortV3,
		Version:    version,
		Confidence: confidence,
	}, true
}

// isSpringBootGradleModule checks if the build.gradle file is a Spring Boot module
func (d *SpringBootDetectorV3) isSpringBootGradleModule(fsys fs.FS, gradlePath string) (core.StackInfo, bool) {
	content, err := fs.ReadFile(fsys, gradlePath)
	if err != nil {
		return core.StackInfo{}, false
	}

	contentStr := string(content)
	signals := 0

	// Check for Spring Boot plugin
	hasPlugin := strings.Contains(contentStr, "org.springframework.boot") &&
		(strings.Contains(contentStr, "id 'org.springframework.boot'") ||
			strings.Contains(contentStr, "id(\"org.springframework.boot\")"))
	if hasPlugin {
		signals++
	}

	// Check for Spring Boot starter dependencies
	hasStarter := strings.Contains(contentStr, "spring-boot-starter-web") ||
		strings.Contains(contentStr, "spring-boot-starter-actuator") ||
		strings.Contains(contentStr, "spring-boot-starter-webflux") ||
		strings.Contains(contentStr, "spring-boot-starter-data-jpa")
	if hasStarter {
		signals++
	}

	// Check for Spring Boot annotations in Java files
	hasAnnotations := d.checkForSpringBootAnnotations(fsys, filepath.Dir(gradlePath))
	if hasAnnotations {
		signals++
	}

	// Check for application properties/yml
	hasConfig := d.checkForSpringBootConfig(fsys, filepath.Dir(gradlePath))
	if hasConfig {
		signals++
	}

	// For Gradle modules, we need either plugin or starter dependencies
	if !hasPlugin && !hasStarter {
		return core.StackInfo{}, false
	}

	// Calculate confidence based on signals
	confidence := 0.5
	if signals >= 3 {
		confidence = 1.0
	} else if signals == 2 {
		confidence = 0.8
	}

	// Extract version if available
	version := ""
	if hasPlugin {
		version = d.extractVersionFromGradle(contentStr)
	}

	return core.StackInfo{
		Name:       "spring-boot",
		BuildTool:  "gradle",
		Port:       defaultPortV3,
		Version:    version,
		Confidence: confidence,
	}, true
}

// extractVersionFromMaven extracts the Spring Boot version from pom.xml
func (d *SpringBootDetectorV3) extractVersionFromMaven(content string) string {
	// Look for version in parent
	if matches := versionRXV3.FindStringSubmatch(content); len(matches) > 1 {
		if matches[1] != "" {
			return matches[1]
		}
		if matches[2] != "" {
			return matches[2]
		}
	}
	return ""
}

// extractVersionFromGradle extracts the Spring Boot version from build.gradle
func (d *SpringBootDetectorV3) extractVersionFromGradle(content string) string {
	// Look for version in plugin declaration
	if matches := versionRXV3.FindStringSubmatch(content); len(matches) > 1 {
		if matches[1] != "" {
			return matches[1]
		}
		if matches[2] != "" {
			return matches[2]
		}
	}
	return ""
}

// checkForSpringBootAnnotations checks for Spring Boot annotations in Java files
func (d *SpringBootDetectorV3) checkForSpringBootAnnotations(fsys fs.FS, dir string) bool {
	var found bool
	fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".java") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return nil
			}
			if strings.Contains(string(content), "@SpringBootApplication") {
				found = true
			}
		}
		return nil
	})
	return found
}

// checkForSpringBootConfig checks for Spring Boot configuration files
func (d *SpringBootDetectorV3) checkForSpringBootConfig(fsys fs.FS, dir string) bool {
	configFiles := []string{
		"src/main/resources/application.properties",
		"src/main/resources/application.yml",
		"src/main/resources/application.yaml",
	}
	for _, file := range configFiles {
		if _, err := fs.ReadFile(fsys, filepath.Join(dir, file)); err == nil {
			return true
		}
	}
	return false
}

// log sends a message to the log sink if available
func (d *SpringBootDetectorV3) log(format string, args ...interface{}) {
	if d.logSink != nil {
		d.logSink.Log(fmt.Sprintf(format, args...))
	}
}

// Name returns the detector name
func (d *SpringBootDetectorV3) Name() string {
	return "spring-boot"
}

// Describe returns a description of what the detector looks for
func (d *SpringBootDetectorV3) Describe() string {
	return "Detects Spring Boot projects by checking for Spring Boot dependencies, plugins, annotations, and configuration files"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetectorV3) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}
