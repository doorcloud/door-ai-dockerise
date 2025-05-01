package spring

import (
	"context"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

const maxDepth = 5

var springGradleRX = regexp.MustCompile(`(?i)org\.springframework\.boot`)

// SpringBootDetectorV2 implements core.Detector for Spring Boot projects
type SpringBootDetectorV2 struct {
	logSink core.LogSink
}

// NewSpringBootDetectorV2 creates a new Spring Boot detector
func NewSpringBootDetectorV2() *SpringBootDetectorV2 {
	return &SpringBootDetectorV2{}
}

// Detect implements the Detector interface
func (d *SpringBootDetectorV2) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink

	// Check for Maven projects
	if info, found, err := d.detectMaven(fsys); err != nil {
		return core.StackInfo{}, false, err
	} else if found {
		return info, true, nil
	}

	// Check for Gradle projects
	if info, found, err := d.detectGradle(fsys); err != nil {
		return core.StackInfo{}, false, err
	} else if found {
		return info, true, nil
	}

	// Check for buildSrc
	if d.dirExists(fsys, "buildSrc") {
		if d.matchBuildSrc(fsys) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"buildSrc/build.gradle.kts"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle source=buildsrc")
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// detectMaven checks for Maven projects (single and multi-module)
func (d *SpringBootDetectorV2) detectMaven(fsys fs.FS) (core.StackInfo, bool, error) {
	// Check for parent pom.xml
	var pomPaths []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Calculate depth by counting path separators
			depth := strings.Count(path, string(filepath.Separator))
			if depth > maxDepth {
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

	// First check for app/pom.xml or application/pom.xml
	for _, pomPath := range pomPaths {
		dir := filepath.Dir(pomPath)
		base := filepath.Base(dir)
		if (base == "app" || base == "application") && d.isSpringBootMavenModule(fsys, pomPath) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				DetectedFiles: []string{pomPath},
			}
			d.log("detector=spring-boot found=true buildtool=maven")
			return info, true, nil
		}
	}

	// Then check other pom.xml files
	for _, pomPath := range pomPaths {
		if d.isSpringBootMavenModule(fsys, pomPath) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "maven",
				DetectedFiles: []string{pomPath},
			}
			d.log("detector=spring-boot found=true buildtool=maven")
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// readFileIfExists reads a file if it exists, returns nil if not found
func readFileIfExists(fsys fs.FS, path string) []byte {
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil
	}
	return content
}

// detectGradle checks for Gradle projects (Groovy, Kotlin, and multi-module)
func (d *SpringBootDetectorV2) detectGradle(fsys fs.FS) (core.StackInfo, bool, error) {
	// Check for build.gradle* and settings.gradle*
	var gradlePaths []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Calculate depth by counting path separators
			depth := strings.Count(path, string(filepath.Separator))
			if depth > maxDepth {
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

	// First check for app/build.gradle or application/build.gradle
	for _, gradlePath := range gradlePaths {
		dir := filepath.Dir(gradlePath)
		base := filepath.Base(dir)
		if (base == "app" || base == "application") && d.isSpringBootGradleModule(fsys, gradlePath) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{gradlePath},
			}
			d.log("detector=spring-boot found=true buildtool=gradle")
			return info, true, nil
		}
	}

	// Then check other build.gradle files
	for _, gradlePath := range gradlePaths {
		if d.isSpringBootGradleModule(fsys, gradlePath) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{gradlePath},
			}
			d.log("detector=spring-boot found=true buildtool=gradle")
			return info, true, nil
		}
	}

	// Check for version catalog
	if content := readFileIfExists(fsys, "gradle/libs.versions.toml"); content != nil {
		if springGradleRX.Match(content) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"gradle/libs.versions.toml"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle source=version-catalog")
			return info, true, nil
		}
	}

	// Check for settings.gradle* with alias
	if settingsContent := readFileIfExists(fsys, "settings.gradle"); settingsContent != nil {
		if strings.Contains(string(settingsContent), "alias(") && springGradleRX.Match(settingsContent) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				DetectedFiles: []string{"settings.gradle"},
			}
			d.log("detector=spring-boot found=true buildtool=gradle source=settings-alias")
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// isSpringBootMavenModule checks if the pom.xml is a Spring Boot module
func (d *SpringBootDetectorV2) isSpringBootMavenModule(fsys fs.FS, pomPath string) bool {
	content, err := fs.ReadFile(fsys, pomPath)
	if err != nil {
		return false
	}

	contentStr := string(content)

	// Must have Spring Boot parent or platform BOM
	hasParent := strings.Contains(contentStr, "<parent>") &&
		(strings.Contains(contentStr, "<groupId>org.springframework.boot</groupId>") ||
			strings.Contains(contentStr, "<groupId>io.spring.platform</groupId>")) &&
		(strings.Contains(contentStr, "<artifactId>spring-boot-starter-parent</artifactId>") ||
			strings.Contains(contentStr, "<artifactId>platform-bom</artifactId>"))

	// Must have Spring Boot plugin
	hasPlugin := strings.Contains(contentStr, "<groupId>org.springframework.boot</groupId>") &&
		strings.Contains(contentStr, "<artifactId>spring-boot-maven-plugin</artifactId>")

	// Must have Spring Boot starter dependencies
	hasStarter := strings.Contains(contentStr, "<artifactId>spring-boot-starter-web</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-actuator</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-webflux</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-data-jpa</artifactId>")

	// For modules, we need either:
	// 1. Spring Boot parent + starter dependencies, or
	// 2. Spring Boot plugin + starter dependencies
	return (hasParent || hasPlugin) && hasStarter
}

// isSpringBootGradleModule checks if the build.gradle file is a Spring Boot module
func (d *SpringBootDetectorV2) isSpringBootGradleModule(fsys fs.FS, gradlePath string) bool {
	content, err := fs.ReadFile(fsys, gradlePath)
	if err != nil {
		return false
	}

	contentStr := string(content)

	// Must have Spring Boot plugin
	hasPlugin := strings.Contains(contentStr, "org.springframework.boot") &&
		(strings.Contains(contentStr, "id 'org.springframework.boot'") ||
			strings.Contains(contentStr, "id(\"org.springframework.boot\")"))

	// Must have Spring Boot starter dependencies
	hasStarter := strings.Contains(contentStr, "spring-boot-starter-web") ||
		strings.Contains(contentStr, "spring-boot-starter-actuator") ||
		strings.Contains(contentStr, "spring-boot-starter-webflux") ||
		strings.Contains(contentStr, "spring-boot-starter-data-jpa")

	// For modules, we need both plugin and starter dependencies
	return hasPlugin && hasStarter
}

// log sends a message to the log sink if available
func (d *SpringBootDetectorV2) log(msg string) {
	if d.logSink != nil {
		d.logSink.Log(msg)
	}
}

// Name returns the detector name
func (d *SpringBootDetectorV2) Name() string {
	return "spring-boot"
}

// Describe returns a description of what the detector looks for
func (d *SpringBootDetectorV2) Describe() string {
	return "Detects Spring Boot projects by checking for Spring Boot dependencies and plugins in Maven or Gradle build files"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetectorV2) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}

// dirExists checks if a directory exists in the filesystem
func (d *SpringBootDetectorV2) dirExists(fsys fs.FS, path string) bool {
	info, err := fs.Stat(fsys, path)
	return err == nil && info.IsDir()
}

// matchBuildSrc checks if the buildSrc directory contains Spring Boot configuration
func (d *SpringBootDetectorV2) matchBuildSrc(fsys fs.FS) bool {
	var found bool
	err := fs.WalkDir(fsys, "buildSrc", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip certain directories
			name := filepath.Base(path)
			if name == "build" || name == ".git" || name == "node_modules" {
				return fs.SkipDir
			}
			// Check depth
			depth := strings.Count(path, string(filepath.Separator))
			if depth > 2 {
				return fs.SkipDir
			}
			return nil
		}
		// Check file content for Spring Boot references
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil
		}
		if springGradleRX.Match(content) {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return err == nil && found
}
