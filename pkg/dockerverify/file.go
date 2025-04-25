package dockerverify

import (
	"os"
	"strings"
)

// defaultShouldCopy returns true if the file should be copied by default
func defaultShouldCopy(p string) bool {
	// Skip hidden files and directories
	if strings.HasPrefix(p, ".") {
		return false
	}

	// Skip test files
	if strings.HasSuffix(p, "_test.go") {
		return false
	}

	// Skip documentation
	if strings.EqualFold(p, "README.md") || strings.EqualFold(p, "LICENSE") {
		return false
	}

	return true
}

// shouldCopy returns true if the file should be copied to the build context
func shouldCopy(p string) bool {
	// Always copy Maven wrapper files
	if strings.HasSuffix(p, "mvnw") || strings.HasSuffix(p, "mvnw.cmd") {
		return true
	}
	if strings.Contains(p, string(os.PathSeparator)+".mvn"+string(os.PathSeparator)) {
		return true
	}

	return defaultShouldCopy(p)
}
