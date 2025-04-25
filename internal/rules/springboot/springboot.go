package springboot

import (
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/rules"
)

// SpringBoot implements the types.Detector interface for Spring Boot projects
type SpringBoot struct{}

func (s *SpringBoot) Name() string {
	return "spring-boot"
}

func (s *SpringBoot) Detect(fsys fs.FS) (bool, error) {
	// Check for pom.xml
	if _, err := fs.Stat(fsys, "pom.xml"); err != nil {
		return false, nil
	}

	// Check for spring-boot dependency in pom.xml
	pomContent, err := fs.ReadFile(fsys, "pom.xml")
	if err != nil {
		return false, nil
	}

	// Simple check for spring-boot dependency
	return containsSpringBootDependency(string(pomContent)), nil
}

func containsSpringBootDependency(pomContent string) bool {
	// This is a simple check - in a real implementation, you'd want to parse the XML properly
	return containsAll(pomContent, "spring-boot", "starter")
}

func containsAll(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

func init() {
	rules.NewRegistry().Register(&SpringBoot{})
}
