package springboot

import (
	"context"
	"encoding/xml"
	"io"
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// SpringBootDetector implements detection rules for Spring Boot projects
type SpringBootDetector struct {
	logSink core.LogSink
	fsys    fs.FS
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
	Parent       *Parent      `xml:"parent"`
	GroupId      string       `xml:"groupId"`
	ArtifactId   string       `xml:"artifactId"`
	Version      string       `xml:"version"`
	Properties   Properties   `xml:"properties"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
}

type Parent struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type Properties struct {
	JavaVersion string `xml:"java.version"`
}

type Dependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink
	d.fsys = fsys

	// Try Maven first
	pomFile := "pom.xml"
	if _, err := fs.Stat(d.fsys, pomFile); err == nil {
		f, err := d.fsys.Open(pomFile)
		if err != nil {
			return core.StackInfo{}, false, err
		}
		defer f.Close()

		var project MavenProject
		if err := xml.NewDecoder(f).Decode(&project); err != nil {
			return core.StackInfo{}, false, err
		}

		// Check if this is a Spring Boot project
		isSpringBoot := d.isSpringBootMavenProject(&project)
		if isSpringBoot {
			version := d.getSpringBootVersion(&project)
			if version == "" {
				// If version not found in current pom, try parent pom
				parentPomPath := "../pom.xml"
				if pf, err := d.fsys.Open(parentPomPath); err == nil {
					defer pf.Close()
					var parentProject MavenProject
					if err := xml.NewDecoder(pf).Decode(&parentProject); err == nil {
						version = d.getSpringBootVersion(&parentProject)
					}
				}
			}

			if version != "" {
				d.logSink.Log("Found Spring Boot Maven project with version " + version)
				return core.StackInfo{
					Name:          "spring-boot",
					Version:       version,
					Confidence:    1.0,
					BuildTool:     "maven",
					DetectedFiles: []string{pomFile},
				}, true, nil
			}
		}
	}

	// Try Gradle next
	gradleFiles := []string{"build.gradle", "build.gradle.kts"}
	for _, gradleFile := range gradleFiles {
		if _, err := fs.Stat(d.fsys, gradleFile); err == nil {
			f, err := d.fsys.Open(gradleFile)
			if err != nil {
				return core.StackInfo{}, false, err
			}
			defer f.Close()

			content, err := io.ReadAll(f)
			if err != nil {
				return core.StackInfo{}, false, err
			}

			if d.isSpringBootGradleProject(string(content)) {
				version := d.extractSpringBootVersion(string(content))
				if version != "" {
					d.logSink.Log("Found Spring Boot Gradle project with version " + version)
					return core.StackInfo{
						Name:          "spring-boot",
						Version:       version,
						Confidence:    1.0,
						BuildTool:     "gradle",
						DetectedFiles: []string{gradleFile},
					}, true, nil
				}
			}
		}
	}

	return core.StackInfo{}, false, nil
}

func (d *SpringBootDetector) isSpringBootMavenProject(project *MavenProject) bool {
	// Check parent
	if project.Parent != nil &&
		project.Parent.GroupId == "org.springframework.boot" &&
		project.Parent.ArtifactId == "spring-boot-starter-parent" {
		return true
	}

	// Check dependencies
	for _, dep := range project.Dependencies {
		if dep.GroupId == "org.springframework.boot" &&
			strings.Contains(dep.ArtifactId, "spring-boot") {
			return true
		}
	}

	return false
}

func (d *SpringBootDetector) getSpringBootVersion(project *MavenProject) string {
	// Check parent version first
	if project.Parent != nil &&
		project.Parent.GroupId == "org.springframework.boot" &&
		project.Parent.ArtifactId == "spring-boot-starter-parent" {
		return project.Parent.Version
	}

	// Check dependencies
	for _, dep := range project.Dependencies {
		if dep.GroupId == "org.springframework.boot" &&
			strings.Contains(dep.ArtifactId, "spring-boot") &&
			dep.Version != "" {
			return dep.Version
		}
	}

	return ""
}

func (d *SpringBootDetector) extractSpringBootVersion(content string) string {
	// Match version in Gradle files
	patterns := []string{
		`id\("org\.springframework\.boot"\)\s+version\s+"([^"]+)"`,
		`id\s+'org\.springframework\.boot'\s+version\s+'([^']+)'`,
		`org\.springframework\.boot:spring-boot-gradle-plugin:([^"'\s]+)`,
		`springBootVersion\s*=\s*['"]([^'"]+)['"]`,
		`spring-boot-gradle-plugin:([^"'\s]+)`,
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

func (d *SpringBootDetector) isSpringBootGradleProject(content string) bool {
	patterns := []string{
		`id\s*['"(]org\.springframework\.boot['")\s]`,
		`org\.springframework\.boot:spring-boot-gradle-plugin`,
		`spring-boot-starter-web`,
		`spring-boot-starter-parent`,
	}

	for _, pattern := range patterns {
		if regexp.MustCompile(pattern).MatchString(content) {
			return true
		}
	}

	return false
}
