package spring

import (
	"encoding/xml"
	"io"
	"io/fs"
	"regexp"
)

// MavenProject represents a Maven project's pom.xml structure
type MavenProject struct {
	Parent               *Parent               `xml:"parent"`
	DependencyManagement *DependencyManagement `xml:"dependencyManagement"`
}

type Parent struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

type DependencyManagement struct {
	Dependencies []Dependency `xml:"dependencies>dependency"`
}

type Dependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// extractSpringBootVersionFromMaven extracts Spring Boot version from Maven project
func extractSpringBootVersionFromMaven(fsys fs.FS) (*string, error) {
	pomFile := "pom.xml"
	if f, err := fsys.Open(pomFile); err == nil {
		defer f.Close()
		var project MavenProject
		if err := xml.NewDecoder(f).Decode(&project); err != nil {
			return nil, err
		}

		// Check parent version
		if project.Parent != nil &&
			project.Parent.GroupId == "org.springframework.boot" &&
			project.Parent.ArtifactId == "spring-boot-starter-parent" {
			return &project.Parent.Version, nil
		}

		// Check dependencyManagement BOM
		if project.DependencyManagement != nil {
			for _, dep := range project.DependencyManagement.Dependencies {
				if dep.GroupId == "org.springframework.boot" &&
					dep.ArtifactId == "spring-boot-dependencies" &&
					dep.Version != "" {
					return &dep.Version, nil
				}
			}
		}
	}

	return nil, nil
}

// extractSpringBootVersionFromGradle extracts Spring Boot version from Gradle project
func extractSpringBootVersionFromGradle(fsys fs.FS) (*string, error) {
	// Try build.gradle and build.gradle.kts
	gradleFiles := []string{"build.gradle", "build.gradle.kts"}
	for _, gradleFile := range gradleFiles {
		if f, err := fsys.Open(gradleFile); err == nil {
			defer f.Close()
			content, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}

			// Try Kotlin DSL
			kotlinRe := regexp.MustCompile(`id\("org\.springframework\.boot"\)\s+version\s+"([^"]+)"`)
			if matches := kotlinRe.FindStringSubmatch(string(content)); len(matches) > 1 {
				return &matches[1], nil
			}

			// Try Groovy DSL
			groovyRe := regexp.MustCompile(`id\s+'org\.springframework\.boot'\s+version\s+'([^']+)'`)
			if matches := groovyRe.FindStringSubmatch(string(content)); len(matches) > 1 {
				return &matches[1], nil
			}
		}
	}

	// Try libs.versions.toml
	if f, err := fsys.Open("gradle/libs.versions.toml"); err == nil {
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		tomlRe := regexp.MustCompile(`spring-boot\s*=\s*"([^"]+)"`)
		if matches := tomlRe.FindStringSubmatch(string(content)); len(matches) > 1 {
			return &matches[1], nil
		}
	}

	return nil, nil
}

// ExtractSpringBootVersion extracts Spring Boot version from the project
func ExtractSpringBootVersion(fsys fs.FS) (*string, error) {
	// Try Maven first
	if _, err := fs.Stat(fsys, "pom.xml"); err == nil {
		return extractSpringBootVersionFromMaven(fsys)
	}

	// Try Gradle
	if _, err := fs.Stat(fsys, "build.gradle"); err == nil {
		return extractSpringBootVersionFromGradle(fsys)
	}
	if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		return extractSpringBootVersionFromGradle(fsys)
	}

	return nil, nil
}
