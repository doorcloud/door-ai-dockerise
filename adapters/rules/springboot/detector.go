package springboot

import (
	"encoding/xml"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// SpringBootDetector implements detection rules for Spring Boot projects
type SpringBootDetector struct {
	logSink core.LogSink
}

// NewSpringBootDetector creates a new Spring Boot detector
func NewSpringBootDetector() *SpringBootDetector {
	return &SpringBootDetector{}
}

// Name returns the detector name
func (d *SpringBootDetector) Name() string {
	return "spring-boot"
}

// Describe returns a description of what the detector looks for
func (d *SpringBootDetector) Describe() string {
	return "Detects Spring Boot projects by checking for Spring Boot dependencies in Maven or Gradle build files"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetector) SetLogSink(sink core.LogSink) {
	d.logSink = sink
}

// MavenProject represents a Maven project's pom.xml structure
type MavenProject struct {
	Parent     Parent `xml:"parent"`
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Properties struct {
		JavaVersion string `xml:"java.version"`
	} `xml:"properties"`
	Dependencies []struct {
		GroupId    string `xml:"groupId"`
		ArtifactId string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"dependencies>dependency"`
}

type Parent struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetector) Detect(fsys core.FS, path string) (core.StackInfo, error) {
	stackInfo := core.StackInfo{
		Name:       "spring-boot",
		Port:       8080,
		Confidence: 0.0,
	}

	// Try Maven first
	pomPath := filepath.Join(path, "pom.xml")
	if _, err := fsys.Stat(pomPath); err == nil {
		file, err := fsys.Open(pomPath)
		if err != nil {
			return stackInfo, err
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return stackInfo, err
		}

		var project MavenProject
		if err := xml.Unmarshal(content, &project); err != nil {
			return stackInfo, err
		}

		// Check for Spring Boot parent
		isSpringBootParent := project.Parent.GroupId == "org.springframework.boot" &&
			project.Parent.ArtifactId == "spring-boot-starter-parent"

		// Check for Spring Boot dependencies
		hasSpringBootDeps := false
		for _, dep := range project.Dependencies {
			if dep.GroupId == "org.springframework.boot" {
				hasSpringBootDeps = true
				break
			}
		}

		if isSpringBootParent || hasSpringBootDeps {
			stackInfo.BuildTool = "maven"
			stackInfo.DetectedFiles = []string{pomPath}

			// Get version from parent if available
			if isSpringBootParent && project.Parent.Version != "" {
				stackInfo.Version = project.Parent.Version
			}

			// Check dependencies for version if not found in parent
			if stackInfo.Version == "" {
				for _, dep := range project.Dependencies {
					if dep.GroupId == "org.springframework.boot" && dep.Version != "" {
						stackInfo.Version = dep.Version
						break
					}
				}
			}

			stackInfo.Confidence = 1.0
			if !isSpringBootParent {
				stackInfo.Confidence = 0.8
			}

			d.logSink.Log("Detected Spring Boot project with version " + stackInfo.Version)
			return stackInfo, nil
		}
	}

	// Try Gradle
	gradleFiles := []string{
		filepath.Join(path, "build.gradle"),
		filepath.Join(path, "build.gradle.kts"),
	}

	for _, gradleFile := range gradleFiles {
		if _, err := fsys.Stat(gradleFile); err == nil {
			file, err := fsys.Open(gradleFile)
			if err != nil {
				continue
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				continue
			}

			contentStr := string(content)
			version := d.extractSpringBootVersion(contentStr)

			if strings.Contains(contentStr, "org.springframework.boot") {
				stackInfo.BuildTool = "gradle"
				stackInfo.DetectedFiles = []string{gradleFile}
				stackInfo.Version = version
				stackInfo.Confidence = 1.0

				if version == "" {
					stackInfo.Confidence = 0.8
				}

				d.logSink.Log("Detected Spring Boot project with version " + stackInfo.Version)
				return stackInfo, nil
			}
		}
	}

	return stackInfo, nil
}

func (d *SpringBootDetector) extractSpringBootVersion(content string) string {
	// Match version in Gradle files
	patterns := []string{
		`id\("org\.springframework\.boot"\)\s+version\s+"([^"]+)"`,
		`org\.springframework\.boot:spring-boot-gradle-plugin:([^"'\s]+)`,
		`springBootVersion\s*=\s*['"]([^'"]+)['"]`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(content); len(match) > 1 {
			return match[1]
		}
	}

	return ""
}

func extractJavaVersion(content string) string {
	// Try to find Java version in sourceCompatibility
	sourcePattern := regexp.MustCompile(`sourceCompatibility\s*=\s*['"]([^'"]+)['"]`)
	if matches := sourcePattern.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
	}

	// Try to find Java version in JavaVersion.VERSION_XX
	versionPattern := regexp.MustCompile(`JavaVersion\.VERSION_(\d+)`)
	if matches := versionPattern.FindStringSubmatch(content); len(matches) > 1 {
		return matches[1]
	}

	return "11" // Default Java version
}

func (d *SpringBootDetector) detectPort(fsys fs.FS) int {
	// Check application.properties
	propFiles := []string{
		"src/main/resources/application.properties",
		"application.properties",
	}
	for _, propFile := range propFiles {
		file, err := fsys.Open(propFile)
		if err == nil {
			defer file.Close()
			content, err := io.ReadAll(file)
			if err == nil {
				portRe := regexp.MustCompile(`server\.port\s*=\s*(\d+)`)
				if matches := portRe.FindStringSubmatch(string(content)); len(matches) > 1 {
					if port, err := strconv.Atoi(matches[1]); err == nil {
						return port
					}
				}
			}
		}
	}

	// Check application.yml
	ymlFiles := []string{
		"src/main/resources/application.yml",
		"application.yml",
		"src/main/resources/application.yaml",
		"application.yaml",
	}
	for _, ymlFile := range ymlFiles {
		file, err := fsys.Open(ymlFile)
		if err == nil {
			defer file.Close()
			content, err := io.ReadAll(file)
			if err == nil {
				portRe := regexp.MustCompile(`server:\s*\n\s*port:\s*(\d+)`)
				if matches := portRe.FindStringSubmatch(string(content)); len(matches) > 1 {
					if port, err := strconv.Atoi(matches[1]); err == nil {
						return port
					}
				}
			}
		}
	}

	return 8080 // Default Spring Boot port
}
