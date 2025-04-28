package springboot

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
	"gopkg.in/yaml.v3"
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

// Detect checks if the given filesystem contains a Spring Boot project
func (d *SpringBootDetector) Detect(ctx context.Context, fsys fs.FS, logSink core.LogSink) (core.StackInfo, bool, error) {
	var buildTool string
	var detectedFiles []string

	// Check for build files
	if pomXml, err := fs.ReadFile(fsys, "pom.xml"); err == nil {
		if strings.Contains(string(pomXml), "spring-boot") {
			buildTool = "maven"
			detectedFiles = append(detectedFiles, "pom.xml")
		}
	}

	if buildTool == "" {
		if gradle, err := fs.ReadFile(fsys, "build.gradle"); err == nil {
			if strings.Contains(string(gradle), "org.springframework.boot") {
				buildTool = "gradle"
				detectedFiles = append(detectedFiles, "build.gradle")
			}
		}
	}

	if buildTool == "" {
		if gradleKts, err := fs.ReadFile(fsys, "build.gradle.kts"); err == nil {
			if strings.Contains(string(gradleKts), "org.springframework.boot") {
				buildTool = "gradle"
				detectedFiles = append(detectedFiles, "build.gradle.kts")
			}
		}
	}

	if buildTool == "" {
		return core.StackInfo{}, false, nil
	}

	// Default port
	port := 8080

	// Check for application properties/yml files
	configFiles := []string{
		"src/main/resources/application.properties",
		"src/main/resources/application.yml",
		"src/main/resources/application.yaml",
	}

	for _, configFile := range configFiles {
		if content, err := fs.ReadFile(fsys, configFile); err == nil {
			detectedFiles = append(detectedFiles, configFile)

			// Try to extract port from properties file
			if strings.HasSuffix(configFile, ".properties") {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "server.port=") {
						portStr := strings.TrimPrefix(line, "server.port=")
						if portStr != "" {
							port = parseInt(portStr)
						}
					}
				}
			} else {
				// Try to extract port from YAML file
				var config struct {
					Server struct {
						Port int `yaml:"port"`
					} `yaml:"server"`
				}
				if err := yaml.Unmarshal(content, &config); err == nil && config.Server.Port != 0 {
					port = config.Server.Port
				}
			}
		}
	}

	if d.logSink != nil {
		d.logSink.Log("detector=springboot found=true")
	}

	return core.StackInfo{
		Name:          "springboot",
		BuildTool:     buildTool,
		Port:          port,
		DetectedFiles: detectedFiles,
	}, true, nil
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
