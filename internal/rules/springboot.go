package rules

import (
	"io/fs"
)

// SpringBoot implements the Rule interface for Spring Boot projects
type SpringBoot struct{}

func (s *SpringBoot) Name() string {
	return "spring-boot"
}

func (s *SpringBoot) Detect(fsys fs.FS) bool {
	// Check for pom.xml
	if _, err := fs.Stat(fsys, "pom.xml"); err != nil {
		return false
	}

	// Check for spring-boot dependency in pom.xml
	pomContent, err := fs.ReadFile(fsys, "pom.xml")
	if err != nil {
		return false
	}

	// Simple check for spring-boot dependency
	return containsSpringBootDependency(string(pomContent))
}

func containsSpringBootDependency(pomContent string) bool {
	// This is a simple check - in a real implementation, you'd want to parse the XML properly
	return containsAll(pomContent, "spring-boot", "starter")
}

func containsAll(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if !contains(s, substr) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

func init() {
	Register(&SpringBoot{})
}
