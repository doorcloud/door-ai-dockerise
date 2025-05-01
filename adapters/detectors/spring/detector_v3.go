package spring

import (
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
	springGradleRXV3          = regexp.MustCompile(`(?i)spring[-_.]?boot`)
	springBootMavenRXV3       = regexp.MustCompile(`(?s)<parent>.*?<groupId>\s*org\.springframework\.boot\s*</groupId>.*?<artifactId>\s*spring-boot-starter-parent\s*</artifactId>.*?<version>\s*(\d+\.\d+\.\d+)(?:-[^<]+)?\s*</version>.*?</parent>`)
	springBootMavenDepRXV3    = regexp.MustCompile(`<dependency>.*?<groupId>\s*org\.springframework\.boot\s*</groupId>.*?<artifactId>\s*spring-boot-starter[^<]*</artifactId>.*?<version>\s*(\d+\.\d+\.\d+)(?:-[^<]+)?\s*</version>.*?</dependency>`)
	springBootMavenPluginRXV3 = regexp.MustCompile(`<plugin>.*?<groupId>\s*org\.springframework\.boot\s*</groupId>.*?<artifactId>\s*spring-boot-maven-plugin\s*</artifactId>.*?<version>\s*(\d+\.\d+\.\d+)(?:-[^<]+)?\s*</version>.*?</plugin>`)
	springBootGradleRXV3      = regexp.MustCompile(`(?:id\s+['"]org\.springframework\.boot['"].*?version\s+['"]|springBootVersion\s*=\s*['"]|org\.springframework\.boot:spring-boot[^:]*:|spring-boot\s*=\s*['"])(\d+\.\d+\.\d+)(?:-[^'"]+)?['"]`)
	springBootAppRXV3         = regexp.MustCompile(`@SpringBootApplication`)
)

// SpringBootDetectorV3 implements core.Detector for Spring Boot projects
type SpringBootDetectorV3 struct {
	logSink core.LogSink
}

// NewSpringBootDetectorV3 creates a new Spring Boot detector
func NewSpringBootDetectorV3() *SpringBootDetectorV3 {
	return &SpringBootDetectorV3{}
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

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetectorV3) Detect(fsys fs.FS) (core.StackInfo, bool) {
	// First, check for Maven projects
	pomFiles, err := fs.Glob(fsys, "pom.xml")
	if err == nil && len(pomFiles) > 0 {
		for _, pomFile := range pomFiles {
			if info, ok := d.isSpringBootMavenModule(fsys, pomFile); ok {
				info.DetectedFiles = []string{pomFile}
				return info, true
			}
		}
	}

	// Then, check for Gradle projects
	gradleFiles, err := fs.Glob(fsys, "build.gradle*")
	if err == nil && len(gradleFiles) > 0 {
		for _, gradleFile := range gradleFiles {
			if info, ok := d.isSpringBootGradleModule(fsys, gradleFile); ok {
				info.DetectedFiles = []string{gradleFile}
				return info, true
			}
		}
	}

	// Check subdirectories for Maven projects
	pomFiles, err = fs.Glob(fsys, "*/pom.xml")
	if err == nil && len(pomFiles) > 0 {
		for _, pomFile := range pomFiles {
			if info, ok := d.isSpringBootMavenModule(fsys, pomFile); ok {
				info.DetectedFiles = []string{pomFile}
				return info, true
			}
		}
	}

	// Check subdirectories for Gradle projects
	gradleFiles, err = fs.Glob(fsys, "*/build.gradle*")
	if err == nil && len(gradleFiles) > 0 {
		for _, gradleFile := range gradleFiles {
			if info, ok := d.isSpringBootGradleModule(fsys, gradleFile); ok {
				info.DetectedFiles = []string{gradleFile}
				return info, true
			}
		}
	}

	// Check deep nested directories for Maven projects
	pomFiles, err = fs.Glob(fsys, "**/pom.xml")
	if err == nil && len(pomFiles) > 0 {
		for _, pomFile := range pomFiles {
			if info, ok := d.isSpringBootMavenModule(fsys, pomFile); ok {
				info.DetectedFiles = []string{pomFile}
				return info, true
			}
		}
	}

	// Check deep nested directories for Gradle projects
	gradleFiles, err = fs.Glob(fsys, "**/build.gradle*")
	if err == nil && len(gradleFiles) > 0 {
		for _, gradleFile := range gradleFiles {
			if info, ok := d.isSpringBootGradleModule(fsys, gradleFile); ok {
				info.DetectedFiles = []string{gradleFile}
				return info, true
			}
		}
	}

	return core.StackInfo{}, false
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

	// Check for version catalog first
	if content := readFileIfExists(fsys, "gradle/libs.versions.toml"); content != nil {
		if springGradleRXV3.Match(content) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          defaultPortV3,
				DetectedFiles: []string{"gradle/libs.versions.toml"},
				Confidence:    1.0,
			}
			if matches := springBootGradleRXV3.FindSubmatch(content); len(matches) > 1 {
				info.Version = string(matches[1])
			}
			d.log("detector=spring-boot found=true buildtool=gradle source=version-catalog")
			return info, true, nil
		}
	}

	// Check for settings.gradle with alias
	if content := readFileIfExists(fsys, "settings.gradle"); content != nil {
		if strings.Contains(string(content), "alias(") && springGradleRXV3.Match(content) {
			info := core.StackInfo{
				Name:          "spring-boot",
				BuildTool:     "gradle",
				Port:          defaultPortV3,
				DetectedFiles: []string{"settings.gradle"},
				Confidence:    1.0,
			}
			if matches := springBootGradleRXV3.FindSubmatch(content); len(matches) > 1 {
				info.Version = string(matches[1])
			}
			d.log("detector=spring-boot found=true buildtool=gradle source=settings-alias")
			return info, true, nil
		}
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

	// First check for app/build.gradle or application/build.gradle
	for _, gradlePath := range gradlePaths {
		dir := filepath.Dir(gradlePath)
		base := filepath.Base(dir)
		if base == "app" || base == "application" {
			if info, found := d.isSpringBootGradleModule(fsys, gradlePath); found {
				info.DetectedFiles = []string{gradlePath}
				d.log("detector=spring-boot found=true buildtool=gradle")
				return info, true, nil
			}
		}
	}

	// Then check other build.gradle files
	for _, gradlePath := range gradlePaths {
		if info, found := d.isSpringBootGradleModule(fsys, gradlePath); found {
			info.DetectedFiles = []string{gradlePath}
			d.log("detector=spring-boot found=true buildtool=gradle")
			return info, true, nil
		}
	}

	return core.StackInfo{}, false, nil
}

// isSpringBootMavenModule checks if the pom.xml file is a Spring Boot module
func (d *SpringBootDetectorV3) isSpringBootMavenModule(fsys fs.FS, pomPath string) (core.StackInfo, bool) {
	content, err := fs.ReadFile(fsys, pomPath)
	if err != nil {
		return core.StackInfo{}, false
	}

	contentStr := string(content)
	signals := 0

	// Check for Spring Boot parent
	hasParent := strings.Contains(contentStr, "<groupId>org.springframework.boot</groupId>") &&
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-parent</artifactId>")
	if hasParent {
		signals += 2
	}

	// Check for Spring Boot starter dependencies
	hasStarter := strings.Contains(contentStr, "<artifactId>spring-boot-starter-web</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-actuator</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-webflux</artifactId>") ||
		strings.Contains(contentStr, "<artifactId>spring-boot-starter-data-jpa</artifactId>")
	if hasStarter {
		signals += 2
	}

	// Check for Spring Boot annotations in Java files
	hasAnnotations := d.checkForSpringBootAnnotations(fsys, filepath.Dir(pomPath))
	if hasAnnotations {
		signals++
	}

	// Check for Spring Boot configuration files
	hasConfig := d.checkForSpringBootConfig(fsys, filepath.Dir(pomPath))
	if hasConfig {
		signals++
	}

	// For Maven modules, we need either parent or starter dependencies
	if !hasParent && !hasStarter {
		return core.StackInfo{}, false
	}

	// Calculate confidence based on signals
	confidence := 0.5
	switch {
	case signals >= 4:
		confidence = 1.0
	case signals >= 2:
		confidence = 0.8
	}

	// Extract version if available
	version := d.extractVersionFromMaven(contentStr)
	if version == "" {
		// Try to find version in parent POMs
		version = d.findVersionInParentPOMs(fsys, filepath.Dir(pomPath))
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
		signals += 2
	}

	// Check for Spring Boot starter dependencies
	hasStarter := strings.Contains(contentStr, "spring-boot-starter-web") ||
		strings.Contains(contentStr, "spring-boot-starter-actuator") ||
		strings.Contains(contentStr, "spring-boot-starter-webflux") ||
		strings.Contains(contentStr, "spring-boot-starter-data-jpa")
	if hasStarter {
		signals += 2
	}

	// Check for Spring Boot annotations in Java files
	hasAnnotations := d.checkForSpringBootAnnotations(fsys, filepath.Dir(gradlePath))
	if hasAnnotations {
		signals++
	}

	// Check for Spring Boot configuration files
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
	switch {
	case signals >= 4:
		confidence = 1.0
	case signals >= 2:
		confidence = 0.8
	}

	// Extract version if available
	version := d.extractVersionFromGradle(contentStr)
	if version == "" {
		// Try to find version in version catalog
		if versionCatalog := readFileIfExists(fsys, "gradle/libs.versions.toml"); versionCatalog != nil {
			if matches := springBootGradleRXV3.FindSubmatch(versionCatalog); len(matches) > 1 {
				version = string(matches[1])
			}
		}
		// Try to find version in settings.gradle
		if version == "" {
			if settingsContent := readFileIfExists(fsys, "settings.gradle"); settingsContent != nil {
				if matches := springBootGradleRXV3.FindSubmatch(settingsContent); len(matches) > 1 {
					version = string(matches[1])
				}
			}
		}
		// Try to find version in settings.gradle.kts
		if version == "" {
			if settingsContent := readFileIfExists(fsys, "settings.gradle.kts"); settingsContent != nil {
				if matches := springBootGradleRXV3.FindSubmatch(settingsContent); len(matches) > 1 {
					version = string(matches[1])
				}
			}
		}
		// Try to find version in parent build.gradle.kts
		if version == "" {
			if parentContent := readFileIfExists(fsys, "../build.gradle.kts"); parentContent != nil {
				if matches := springBootGradleRXV3.FindSubmatch(parentContent); len(matches) > 1 {
					version = string(matches[1])
				}
			}
		}
	}

	return core.StackInfo{
		Name:       "spring-boot",
		BuildTool:  "gradle",
		Port:       defaultPortV3,
		Version:    version,
		Confidence: confidence,
	}, true
}

// findVersionInParentPOMs recursively searches for Spring Boot version in parent POMs
func (d *SpringBootDetectorV3) findVersionInParentPOMs(fsys fs.FS, dir string) string {
	// Try to find parent pom.xml
	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		return ""
	}

	parentPomPath := filepath.Join(parentDir, "pom.xml")
	content, err := fs.ReadFile(fsys, parentPomPath)
	if err != nil {
		return ""
	}

	version := d.extractVersionFromMaven(string(content))
	if version != "" {
		return version
	}

	// Recursively check parent
	return d.findVersionInParentPOMs(fsys, parentDir)
}

// extractVersionFromMaven extracts Spring Boot version from Maven POM
func (d *SpringBootDetectorV3) extractVersionFromMaven(content string) string {
	// First try parent version
	if matches := springBootMavenRXV3.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
	}

	// Then try dependency version
	if matches := springBootMavenDepRXV3.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
	}

	// Finally try plugin version
	if matches := springBootMavenPluginRXV3.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractVersionFromGradle extracts Spring Boot version from Gradle build file
func (d *SpringBootDetectorV3) extractVersionFromGradle(content string) string {
	if matches := springBootGradleRXV3.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
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
		if strings.HasSuffix(path, ".java") || strings.HasSuffix(path, ".kt") {
			content, err := fs.ReadFile(fsys, path)
			if err != nil {
				return nil
			}
			if springBootAppRXV3.Match(content) {
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
