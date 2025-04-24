package springboot

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/doorcloud/door-ai-dockerise/internal/rules"
)

// Detect returns a Spring Boot rule if the repository contains Spring Boot files.
func Detect(repo string) (*rules.StackRule, error) {
	// Load rule configs
	configs, err := rules.LoadRuleConfigs("internal/rules/configs")
	if err != nil {
		return nil, fmt.Errorf("load rule configs: %w", err)
	}

	config, ok := configs["springboot"]
	if !ok {
		return nil, fmt.Errorf("springboot config not found")
	}

	// Check for build files
	buildFiles := []string{
		"pom.xml",
		"build.gradle*",
		"settings.gradle*",
		"mvnw",
		"gradlew",
	}
	foundBuildFile := false
	for _, file := range buildFiles {
		if _, err := os.Stat(file); err == nil {
			foundBuildFile = true
			break
		}
	}
	if !foundBuildFile {
		return nil, nil
	}

	// Look inside build files for spring-boot plugin/starter
	springBootPattern := regexp.MustCompile(`(?i)spring-boot[-.]`)
	matches, err := filepath.Glob("**/pom.xml")
	if err != nil {
		return nil, fmt.Errorf("glob pom.xml: %w", err)
	}
	matches2, err := filepath.Glob("**/build.gradle*")
	if err != nil {
		return nil, fmt.Errorf("glob build.gradle: %w", err)
	}
	matches = append(matches, matches2...)

	for _, match := range matches {
		content, err := os.ReadFile(match)
		if err != nil {
			continue
		}
		if springBootPattern.Match(content) {
			return createStackRule(config), nil
		}
	}

	// Look for Application class with @SpringBootApplication
	appPattern := regexp.MustCompile(`@SpringBootApplication`)
	matches, err = filepath.Glob("src/main/java/**/*Application.java")
	if err != nil {
		return nil, fmt.Errorf("glob Application.java: %w", err)
	}
	matches2, err = filepath.Glob("src/main/kotlin/**/*Application.kt")
	if err != nil {
		return nil, fmt.Errorf("glob Application.kt: %w", err)
	}
	matches = append(matches, matches2...)

	for _, match := range matches {
		content, err := os.ReadFile(match)
		if err != nil {
			continue
		}
		if appPattern.Match(content) {
			return createStackRule(config), nil
		}
	}

	return nil, nil
}

func createStackRule(config *rules.RuleConfig) *rules.StackRule {
	return &rules.StackRule{
		Name:          config.Name,
		Signatures:    config.Signatures,
		ManifestGlobs: config.ManifestGlobs,
		CodeGlobs:     config.CodeGlobs,
		MainRegex:     config.MainRegex,
		BuildHints:    config.BuildHints,
	}
}

func init() {
	rules.RegisterDetector(Detect)
}
