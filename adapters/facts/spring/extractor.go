package spring

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Spec represents Spring Boot project configuration
type Spec struct {
	BuildTool         string `yaml:"build_tool"`
	JDKVersion        string `yaml:"jdk_version"`
	SpringBootVersion string `yaml:"spring_boot_version"`
	BuildCmd          string `yaml:"build_cmd"`
	Artifact          string `yaml:"artifact"`
	HealthEndpoint    string `yaml:"health_endpoint"`
	Ports             []int  `yaml:"ports"`
}

// Extractor extracts Spring Boot project facts
type Extractor struct{}

// NewExtractor creates a new Spring Boot fact extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extracts Spring Boot project facts from the given path
func (e *Extractor) Extract(path string) (*Spec, error) {
	fsys := os.DirFS(path)

	// Determine build tool
	buildTool, err := e.detectBuildTool(fsys)
	if err != nil {
		return nil, fmt.Errorf("failed to detect build tool: %w", err)
	}

	// Extract build file content
	buildFile, err := e.readBuildFile(fsys, buildTool)
	if err != nil {
		return nil, fmt.Errorf("failed to read build file: %w", err)
	}

	// Extract facts from build file
	spec := &Spec{
		BuildTool:      buildTool,
		HealthEndpoint: "/actuator/health",
		Ports:          []int{8080},
	}

	switch buildTool {
	case "maven":
		if err := e.extractMavenFacts(buildFile, spec); err != nil {
			return nil, err
		}
	case "gradle":
		if err := e.extractGradleFacts(buildFile, spec); err != nil {
			return nil, err
		}
	}

	return spec, nil
}

// detectBuildTool detects the build tool used in the project
func (e *Extractor) detectBuildTool(fsys fs.FS) (string, error) {
	// Check for Maven
	if _, err := fs.Stat(fsys, "pom.xml"); err == nil {
		return "maven", nil
	}

	// Check for Gradle
	if _, err := fs.Stat(fsys, "build.gradle"); err == nil {
		return "gradle", nil
	}
	if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		return "gradle", nil
	}

	return "", fmt.Errorf("no build tool detected")
}

// readBuildFile reads the build file content
func (e *Extractor) readBuildFile(fsys fs.FS, buildTool string) (string, error) {
	var buildFile string
	switch buildTool {
	case "maven":
		buildFile = "pom.xml"
	case "gradle":
		buildFile = "build.gradle"
		if _, err := fs.Stat(fsys, buildFile); err != nil {
			buildFile = "build.gradle.kts"
		}
	}

	content, err := fs.ReadFile(fsys, buildFile)
	if err != nil {
		return "", fmt.Errorf("failed to read build file: %w", err)
	}

	return string(content), nil
}

// extractMavenFacts extracts facts from Maven build file
func (e *Extractor) extractMavenFacts(content string, spec *Spec) error {
	// Extract Spring Boot version
	if idx := strings.Index(content, "<spring-boot.version>"); idx != -1 {
		start := idx + len("<spring-boot.version>")
		end := strings.Index(content[start:], "<")
		if end != -1 {
			spec.SpringBootVersion = content[start : start+end]
		}
	}

	// Extract Java version
	if idx := strings.Index(content, "<java.version>"); idx != -1 {
		start := idx + len("<java.version>")
		end := strings.Index(content[start:], "<")
		if end != -1 {
			spec.JDKVersion = content[start : start+end]
		}
	}

	// Set build command and artifact
	spec.BuildCmd = "mvn clean package -DskipTests"
	spec.Artifact = "target/*.jar"

	return nil
}

// extractGradleFacts extracts facts from Gradle build file
func (e *Extractor) extractGradleFacts(content string, spec *Spec) error {
	// Extract Spring Boot version
	// Try different patterns for Spring Boot version
	patterns := []string{
		"springBootVersion = '",
		"id 'org.springframework.boot' version '",
		"id(\"org.springframework.boot\") version \"",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(content, pattern); idx != -1 {
			start := idx + len(pattern)
			end := strings.Index(content[start:], "'")
			if end == -1 {
				end = strings.Index(content[start:], "\"")
			}
			if end != -1 {
				spec.SpringBootVersion = content[start : start+end]
				break
			}
		}
	}

	// Extract Java version
	if idx := strings.Index(content, "sourceCompatibility = '"); idx != -1 {
		start := idx + len("sourceCompatibility = '")
		end := strings.Index(content[start:], "'")
		if end != -1 {
			spec.JDKVersion = content[start : start+end]
		}
	}

	// Set build command and artifact
	spec.BuildCmd = "./gradlew build -x test"
	spec.Artifact = "build/libs/*.jar"

	return nil
}
