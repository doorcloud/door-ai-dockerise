package rules

import (
	"os"
	"path/filepath"
	"strings"
)

// springBootRule implements Rule for Spring Boot projects
type springBootRule struct{}

func init() {
	Register(&springBootRule{})
}

func (r *springBootRule) Name() string {
	return "spring-boot"
}

func (r *springBootRule) Detect(repo string) bool {
	// Check for Maven
	if _, err := os.Stat(filepath.Join(repo, "pom.xml")); err == nil {
		return true
	}

	// Check for Gradle
	if _, err := os.Stat(filepath.Join(repo, "gradlew")); err == nil {
		return true
	}

	// Check for Spring Boot application class
	return r.hasSpringBootApp(repo)
}

func (r *springBootRule) Facts(repo string) map[string]any {
	return map[string]any{
		"framework": "Spring Boot",
		"language":  "Java",
	}
}

func (r *springBootRule) hasSpringBootApp(repo string) bool {
	var found bool
	filepath.Walk(repo, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".java") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), "@SpringBootApplication") {
			found = true
			return filepath.SkipDir
		}
		return nil
	})
	return found
}
