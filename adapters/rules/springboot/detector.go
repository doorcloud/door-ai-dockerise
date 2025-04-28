package springboot

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
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
	return "springboot"
}

// Describe returns a description of what the detector looks for
func (d *SpringBootDetector) Describe() string {
	return "Detects Spring Boot projects by checking for Spring Boot dependencies in Maven or Gradle build files"
}

// SetLogSink sets the log sink for the detector
func (d *SpringBootDetector) SetLogSink(logSink core.LogSink) {
	d.logSink = logSink
}

type MavenProject struct {
	XMLName    xml.Name `xml:"project"`
	Parent     Parent   `xml:"parent"`
	Properties struct {
		JavaVersion string `xml:"java.version"`
	} `xml:"properties"`
}

type Parent struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	d.logSink = logSink

	// Check for pom.xml first
	pomFile, err := fsys.Open("pom.xml")
	if err == nil {
		defer pomFile.Close()

		var project MavenProject
		if err := xml.NewDecoder(pomFile).Decode(&project); err != nil {
			return core.StackInfo{}, false, fmt.Errorf("failed to parse pom.xml: %w", err)
		}

		// Check if it's a Spring Boot project
		if project.Parent.ArtifactId == "spring-boot-starter-parent" {
			d.logSink.Log(fmt.Sprintf("Detected Spring Boot project with version %s and Java version %s",
				project.Parent.Version, project.Properties.JavaVersion))

			return core.StackInfo{
				Name:          "springboot",
				BuildTool:     "maven",
				Version:       project.Parent.Version,
				DetectedFiles: []string{"pom.xml"},
			}, true, nil
		}
	}

	// Check for Gradle files
	gradleFiles := []string{"build.gradle", "build.gradle.kts"}
	for _, gradleFile := range gradleFiles {
		file, err := fsys.Open(gradleFile)
		if err == nil {
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				return core.StackInfo{}, false, fmt.Errorf("failed to read %s: %w", gradleFile, err)
			}

			contentStr := string(content)
			if strings.Contains(contentStr, "org.springframework.boot") {
				// Try to extract Spring Boot version
				version := "unknown"
				if idx := strings.Index(contentStr, "springBootVersion"); idx != -1 {
					endIdx := strings.Index(contentStr[idx:], "\n")
					if endIdx != -1 {
						versionLine := contentStr[idx : idx+endIdx]
						if strings.Contains(versionLine, "=") {
							version = strings.TrimSpace(strings.Split(versionLine, "=")[1])
							version = strings.Trim(version, "'\"")
						}
					}
				}

				d.logSink.Log(fmt.Sprintf("Detected Spring Boot project with version %s in %s", version, gradleFile))

				return core.StackInfo{
					Name:          "springboot",
					BuildTool:     "gradle",
					Version:       version,
					DetectedFiles: []string{gradleFile},
				}, true, nil
			}
		}
	}

	return core.StackInfo{}, false, nil
}

func parseInt(s string) int {
	var result int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		} else {
			break
		}
	}
	return result
}
