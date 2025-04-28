package spring

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

// Spec represents Spring Boot project configuration
type Spec struct {
	BuildTool         string  `yaml:"build_tool"`
	JDKVersion        string  `yaml:"jdk_version"`
	SpringBootVersion *string `yaml:"spring_boot_version,omitempty"`
	BuildCmd          string  `yaml:"build_cmd"`
	Artifact          string  `yaml:"artifact"`
	HealthEndpoint    string  `yaml:"health_endpoint"`
	Ports             []int   `yaml:"ports"`
	Metadata          map[string]string
	JavaVersion       string `yaml:"java_version"`
}

// Extractor extracts Spring Boot project facts
type Extractor struct{}

// NewExtractor creates a new Spring Boot fact extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extracts Spring Boot project facts from the given path
func (e *Extractor) Extract(fsys fs.FS) (*Spec, error) {
	spec := &Spec{}

	// Extract Java version
	javaVersion, err := e.extractJavaVersion(fsys)
	if err != nil {
		return nil, err
	}
	spec.JavaVersion = javaVersion

	// Extract Spring Boot version
	springBootVersion, err := e.extractSpringBootVersion(fsys)
	if err != nil {
		return nil, err
	}
	if springBootVersion != "" {
		spec.SpringBootVersion = &springBootVersion
	}

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
	spec.BuildTool = buildTool

	// Extract ports and health endpoint
	ports, healthEndpoint, err := e.detectPortsAndHealth(fsys)
	if err != nil {
		return nil, fmt.Errorf("failed to detect ports and health: %w", err)
	}
	spec.Ports = ports
	spec.HealthEndpoint = healthEndpoint

	switch buildTool {
	case "maven":
		if err := e.extractMavenFacts(fsys, buildFile, spec); err != nil {
			return nil, err
		}
	case "gradle":
		if err := e.extractGradleFacts(fsys, buildFile, spec); err != nil {
			return nil, err
		}
	}

	// Try to extract JDK version from toolchain files if not found in build file
	if spec.JDKVersion == "" {
		if jdkVersion, err := e.extractJDKVersionFromToolchain(fsys, buildTool); err == nil && jdkVersion != "" {
			spec.JDKVersion = jdkVersion
		}
	}

	// Detect artifact path
	if artifactPath, err := e.detectArtifactPath(fsys, buildTool, buildFile); err == nil {
		spec.Artifact = artifactPath
	}

	// Detect SBOM path
	if sbomPath, err := e.detectSBOMPath(fsys, buildTool); err == nil && sbomPath != "" {
		if spec.Metadata == nil {
			spec.Metadata = make(map[string]string)
		}
		spec.Metadata["sbom_path"] = sbomPath
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

// extractJDKVersionFromToolchain extracts JDK version from toolchain files
func (e *Extractor) extractJDKVersionFromToolchain(fsys fs.FS, buildTool string) (string, error) {
	switch buildTool {
	case "maven":
		// Try .mvn/toolchains.xml
		if content, err := fs.ReadFile(fsys, ".mvn/toolchains.xml"); err == nil {
			// Look for <release> tag
			if idx := strings.Index(string(content), "<release>"); idx != -1 {
				start := idx + len("<release>")
				end := strings.Index(string(content[start:]), "<")
				if end != -1 {
					return string(content[start : start+end]), nil
				}
			}
		}
	case "gradle":
		// Try Gradle KTS toolchain block
		if content, err := fs.ReadFile(fsys, "build.gradle.kts"); err == nil {
			// Look for toolchain block
			if idx := strings.Index(string(content), "toolchain"); idx != -1 {
				// Look for languageVersion
				if langIdx := strings.Index(string(content[idx:]), "languageVersion.set(JavaLanguageVersion.of("); langIdx != -1 {
					start := idx + langIdx + len("languageVersion.set(JavaLanguageVersion.of(")
					end := strings.Index(string(content[start:]), ")")
					if end != -1 {
						return string(content[start : start+end]), nil
					}
				}
			}
		}
	}
	return "", nil
}

// extractSpringBootVersion extracts Spring Boot version from various sources
func (e *Extractor) extractSpringBootVersion(fsys fs.FS) (string, error) {
	version, err := ExtractSpringBootVersion(fsys)
	if err != nil {
		return "", err
	}
	if version != nil {
		return *version, nil
	}
	return "", nil
}

// detectBuildCommand detects the appropriate build command
func (e *Extractor) detectBuildCommand(fsys fs.FS, buildTool string, artifactPath string) string {
	switch buildTool {
	case "maven":
		if artifactPath != "" && !strings.HasPrefix(artifactPath, "target/") {
			// If artifact is in a module, use -pl for that module
			module := filepath.Dir(artifactPath)
			return fmt.Sprintf("mvn clean package -pl %s -DskipTests", module)
		}
		return "mvn clean package -DskipTests"
	case "gradle":
		// Check if gradlew exists
		if _, err := fs.Stat(fsys, "gradlew"); err == nil {
			if artifactPath != "" && !strings.HasPrefix(artifactPath, "build/") {
				// If artifact is in a module, prefix with module name
				module := filepath.Dir(artifactPath)
				return fmt.Sprintf("./gradlew %s:build -x test", module)
			}
			return "./gradlew build -x test"
		}
		if artifactPath != "" && !strings.HasPrefix(artifactPath, "build/") {
			// If artifact is in a module, prefix with module name
			module := filepath.Dir(artifactPath)
			return fmt.Sprintf("gradle %s:build -x test", module)
		}
		return "gradle build -x test"
	default:
		return ""
	}
}

// extractMavenFacts extracts facts from Maven build file
func (e *Extractor) extractMavenFacts(fsys fs.FS, content string, spec *Spec) error {
	// Extract Spring Boot version
	if version, err := e.extractSpringBootVersion(fsys); err == nil && version != "" {
		spec.SpringBootVersion = &version
	}

	// Extract Java version
	if idx := strings.Index(content, "<java.version>"); idx != -1 {
		start := idx + len("<java.version>")
		end := strings.Index(content[start:], "<")
		if end != -1 {
			spec.JDKVersion = content[start : start+end]
		}
	}

	// Set artifact
	spec.Artifact = "target/*.jar"

	// Set build command
	if cmd := e.detectBuildCommand(fsys, "maven", spec.Artifact); cmd != "" {
		spec.BuildCmd = cmd
	} else {
		spec.BuildCmd = "mvn clean package -DskipTests"
	}

	return nil
}

// extractGradleFacts extracts facts from Gradle build file
func (e *Extractor) extractGradleFacts(fsys fs.FS, content string, spec *Spec) error {
	// Extract Spring Boot version
	if version, err := e.extractSpringBootVersion(fsys); err == nil && version != "" {
		spec.SpringBootVersion = &version
	}

	// Extract Java version
	if idx := strings.Index(content, "sourceCompatibility = '"); idx != -1 {
		start := idx + len("sourceCompatibility = '")
		end := strings.Index(content[start:], "'")
		if end != -1 {
			spec.JDKVersion = content[start : start+end]
		}
	}

	// Set artifact
	spec.Artifact = "build/libs/*.jar"

	// Set build command
	if cmd := e.detectBuildCommand(fsys, "gradle", spec.Artifact); cmd != "" {
		spec.BuildCmd = cmd
	} else {
		spec.BuildCmd = "./gradlew build -x test"
	}

	return nil
}

// detectArtifactPath detects the artifact path for the project
func (e *Extractor) detectArtifactPath(fsys fs.FS, buildTool string, content string) (string, error) {
	switch buildTool {
	case "maven":
		// Check for packaging type
		if idx := strings.Index(content, "<packaging>"); idx != -1 {
			start := idx + len("<packaging>")
			end := strings.Index(content[start:], "<")
			if end != -1 {
				packaging := content[start : start+end]
				if packaging == "war" {
					return "target/*.war", nil
				}
			}
		}
		return "target/*.jar", nil

	case "gradle":
		// Check for multi-module project
		if strings.Contains(content, "include(") || strings.Contains(content, "include ") {
			// Look for application plugin
			if strings.Contains(content, "application") || strings.Contains(content, "org.springframework.boot") {
				// Find the module name
				module := ""
				if idx := strings.Index(content, "rootProject.name = "); idx != -1 {
					start := idx + len("rootProject.name = ")
					end := strings.Index(content[start:], "\n")
					if end != -1 {
						module = strings.Trim(content[start:start+end], " '\"")
					}
				}
				if module != "" {
					return fmt.Sprintf("%s/build/libs/*.jar", module), nil
				}
			}
		}
		return "build/libs/*.jar", nil
	}
	return "", fmt.Errorf("unsupported build tool: %s", buildTool)
}

// detectPortsAndHealth detects ports and health endpoint from application properties
func (e *Extractor) detectPortsAndHealth(fsys fs.FS) ([]int, string, error) {
	// Default values
	ports := []int{8080}
	healthEndpoint := ""

	// Check for Spring Boot parent POM or actuator dependency
	hasActuator := false
	hasSpringBootParent := false
	hasWebOnly := false

	if content, err := fs.ReadFile(fsys, "pom.xml"); err == nil {
		contentStr := string(content)
		hasActuator = strings.Contains(contentStr, "spring-boot-starter-actuator")
		hasSpringBootParent = strings.Contains(contentStr, "<artifactId>spring-boot-starter-parent</artifactId>")
		hasWebOnly = strings.Contains(contentStr, "spring-boot-starter-web") && !strings.Contains(contentStr, "spring-boot-starter-actuator")

		// Check for WAR packaging
		if idx := strings.Index(contentStr, "<packaging>"); idx != -1 {
			start := idx + len("<packaging>")
			end := strings.Index(contentStr[start:], "<")
			if end != -1 {
				packaging := contentStr[start : start+end]
				if packaging == "war" {
					hasActuator = true // WAR packaging implies actuator
				}
			}
		}
	}
	if content, err := fs.ReadFile(fsys, "build.gradle"); err == nil {
		hasActuator = hasActuator || strings.Contains(string(content), "spring-boot-starter-actuator")
		hasSpringBootParent = hasSpringBootParent || strings.Contains(string(content), "org.springframework.boot")
	}
	if content, err := fs.ReadFile(fsys, "build.gradle.kts"); err == nil {
		hasActuator = hasActuator || strings.Contains(string(content), "spring-boot-starter-actuator")
		hasSpringBootParent = hasSpringBootParent || strings.Contains(string(content), "org.springframework.boot")
	}

	// Set health endpoint if actuator is present or if using Spring Boot parent without explicit web-only dependency
	if hasActuator || (hasSpringBootParent && !hasWebOnly) {
		healthEndpoint = "/actuator/health"
	}

	// Look for application properties files
	propFiles := []string{
		"application.properties",
		"application.yml",
		"application.yaml",
		"application.json",
	}

	for _, file := range propFiles {
		if content, err := fs.ReadFile(fsys, file); err == nil {
			// Look for server.port
			if idx := strings.Index(string(content), "server.port="); idx != -1 {
				start := idx + len("server.port=")
				end := strings.Index(string(content[start:]), "\n")
				if end != -1 {
					port := strings.TrimSpace(string(content[start : start+end]))
					if portNum, err := strconv.Atoi(port); err == nil {
						ports = []int{portNum}
					}
				}
			}

			// Look for management.endpoints.web.base-path
			if idx := strings.Index(string(content), "management.endpoints.web.base-path="); idx != -1 {
				start := idx + len("management.endpoints.web.base-path=")
				end := strings.Index(string(content[start:]), "\n")
				if end != -1 {
					basePath := strings.TrimSpace(string(content[start : start+end]))
					healthEndpoint = fmt.Sprintf("%s/health", basePath)
				}
			}
		}
	}

	return ports, healthEndpoint, nil
}

// detectSBOMPath detects the SBOM file path
func (e *Extractor) detectSBOMPath(fsys fs.FS, buildTool string) (string, error) {
	var sbomPaths []string
	if buildTool == "maven" {
		sbomPaths = []string{
			filepath.Join("target", "bom.cdx.json"),
			filepath.Join("target", "*.cdx.json"),
		}
	} else {
		sbomPaths = []string{
			filepath.Join("build", "reports", "bom.cdx.json"),
		}
	}

	for _, pattern := range sbomPaths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	return "", nil
}

// extractJavaVersion extracts Java version from various sources
func (e *Extractor) extractJavaVersion(fsys fs.FS) (string, error) {
	return ExtractJavaVersion(fsys)
}
