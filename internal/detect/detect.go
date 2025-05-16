package detect

import (
	"os"
	"path/filepath"
	"strings"
)

// Detect returns "spring" if the repo looks like a Spring Boot project.
func Detect(repo string) string {
	abs, _ := filepath.Abs(repo)

	if exists(abs, "pom.xml") || exists(abs, "build.gradle") {
		// extra check for spring-boot keyword (optional safety)
		if contains(abs, "pom.xml", "spring-boot") ||
			contains(abs, "build.gradle", "spring-boot") {
			return "spring"
		}
		return "spring"
	}
	return ""
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func contains(dir, name, substr string) bool {
	data, err := os.ReadFile(filepath.Join(dir, name))
	return err == nil && strings.Contains(string(data), substr)
}
