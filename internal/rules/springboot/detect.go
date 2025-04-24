package springboot

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/doorcloud/door-ai-dockerise/internal/rules"
)

// Detect returns a Spring Boot rule if the repository contains Spring Boot files.
func Detect(repo string) (*rules.StackRule, error) {
	// Try to load rule configs
	var config *rules.RuleConfig
	configs, err := rules.LoadRuleConfigs("internal/rules/configs")
	if err == nil {
		if c, ok := configs["springboot"]; ok {
			config = c
		}
	}

	// Use default config if loading failed
	if config == nil {
		config = &rules.RuleConfig{
			Name: "springboot",
			Signatures: []string{
				"**/pom.xml",
				"**/build.gradle",
				"**/application*.{yml,yaml,properties}",
				"**/*Application.java",
			},
		}
	}

	// Check for Maven Spring Boot
	mavenFiles, err := doublestar.Glob(os.DirFS(repo), "**/pom.xml")
	if err != nil {
		return nil, fmt.Errorf("glob pom.xml: %w", err)
	}
	for _, file := range mavenFiles {
		content, err := os.ReadFile(filepath.Join(repo, file))
		if err != nil {
			continue
		}
		// Check for Spring Boot parent or dependency
		if regexp.MustCompile(`<parent>\s*<groupId>org\.springframework\.boot</groupId>`).Match(content) ||
			regexp.MustCompile(`<artifactId>spring-boot-[^<]+</artifactId>`).Match(content) ||
			regexp.MustCompile(`<groupId>org\.springframework\.boot</groupId>`).Match(content) {
			return createStackRule(config), nil
		}
	}

	// Check for Gradle Spring Boot
	gradleFiles, err := doublestar.Glob(os.DirFS(repo), "**/build.gradle*")
	if err != nil {
		return nil, fmt.Errorf("glob build.gradle: %w", err)
	}
	for _, file := range gradleFiles {
		content, err := os.ReadFile(filepath.Join(repo, file))
		if err != nil {
			continue
		}
		// Check for Spring Boot plugin in both Groovy and Kotlin DSL
		if regexp.MustCompile(`id\s*['"](org\.springframework\.boot)['"]`).Match(content) ||
			regexp.MustCompile(`id\s*\(\s*["'](org\.springframework\.boot)["']\s*\)`).Match(content) {
			return createStackRule(config), nil
		}
	}

	// Look for Application class with @SpringBootApplication
	javaFiles, err := doublestar.Glob(os.DirFS(repo), "src/main/java/**/*Application.java")
	if err != nil {
		return nil, fmt.Errorf("glob Application.java: %w", err)
	}
	kotlinFiles, err := doublestar.Glob(os.DirFS(repo), "src/main/kotlin/**/*Application.kt")
	if err != nil {
		return nil, fmt.Errorf("glob Application.kt: %w", err)
	}
	appFiles := append(javaFiles, kotlinFiles...)

	for _, file := range appFiles {
		content, err := os.ReadFile(filepath.Join(repo, file))
		if err != nil {
			continue
		}
		if regexp.MustCompile(`@SpringBootApplication`).Match(content) {
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
