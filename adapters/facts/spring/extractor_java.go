package spring

import (
	"encoding/xml"
	"io"
	"io/fs"
	"regexp"
)

// MavenToolchain represents the structure of Maven toolchains.xml
type MavenToolchain struct {
	XMLName   xml.Name `xml:"toolchains"`
	Toolchain []struct {
		Type     string `xml:"type"`
		Provides struct {
			Version string `xml:"version"`
		} `xml:"provides"`
	} `xml:"toolchain"`
}

// extractJavaVersionFromMaven extracts Java version from Maven project
func extractJavaVersionFromMaven(fsys fs.FS) (string, error) {
	// First try pom.xml
	pomFile := "pom.xml"
	if f, err := fsys.Open(pomFile); err == nil {
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			return "", err
		}

		// Try maven.compiler.release
		releaseRe := regexp.MustCompile(`<maven\.compiler\.release>(\d+)</maven\.compiler\.release>`)
		if matches := releaseRe.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], nil
		}

		// Try maven.compiler.target
		targetRe := regexp.MustCompile(`<maven\.compiler\.target>(\d+)</maven\.compiler\.target>`)
		if matches := targetRe.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], nil
		}

		// Try parent properties
		parentRe := regexp.MustCompile(`<parent>.*?<properties>.*?<java\.version>(\d+)</java\.version>.*?</properties>.*?</parent>`)
		if matches := parentRe.FindStringSubmatch(string(content)); len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Try toolchains.xml
	toolchainFile := ".mvn/toolchains.xml"
	if f, err := fsys.Open(toolchainFile); err == nil {
		defer f.Close()
		var toolchain MavenToolchain
		if err := xml.NewDecoder(f).Decode(&toolchain); err == nil {
			for _, t := range toolchain.Toolchain {
				if t.Type == "jdk" && t.Provides.Version != "" {
					return t.Provides.Version, nil
				}
			}
		}
	}

	return "17", nil // Default fallback
}

// extractJavaVersionFromGradle extracts Java version from Gradle project
func extractJavaVersionFromGradle(fsys fs.FS) (string, error) {
	gradleFiles := []string{"build.gradle", "build.gradle.kts"}
	for _, gradleFile := range gradleFiles {
		if f, err := fsys.Open(gradleFile); err == nil {
			defer f.Close()
			content, err := io.ReadAll(f)
			if err != nil {
				return "", err
			}

			// Try Kotlin DSL
			kotlinRe := regexp.MustCompile(`toolchain\.languageVersion\.set\(JavaLanguageVersion\.of\((\d+)\)\)`)
			if matches := kotlinRe.FindStringSubmatch(string(content)); len(matches) > 1 {
				return matches[1], nil
			}

			// Try Groovy DSL
			groovyRe := regexp.MustCompile(`sourceCompatibility\s*=\s*['"](\d+)['"]`)
			if matches := groovyRe.FindStringSubmatch(string(content)); len(matches) > 1 {
				return matches[1], nil
			}
		}
	}

	return "17", nil // Default fallback
}

// ExtractJavaVersion extracts Java version from the project
func ExtractJavaVersion(fsys fs.FS) (string, error) {
	// Try Maven first
	if _, err := fs.Stat(fsys, "pom.xml"); err == nil {
		return extractJavaVersionFromMaven(fsys)
	}

	// Try Gradle
	if _, err := fs.Stat(fsys, "build.gradle"); err == nil {
		return extractJavaVersionFromGradle(fsys)
	}
	if _, err := fs.Stat(fsys, "build.gradle.kts"); err == nil {
		return extractJavaVersionFromGradle(fsys)
	}

	return "17", nil // Default fallback
}
