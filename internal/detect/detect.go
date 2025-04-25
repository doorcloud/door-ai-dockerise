package detect

import (
	"io/fs"
)

// Rule represents a technology stack detection rule
type Rule struct {
	Name string // e.g. "spring-boot"
	Tool string // e.g. "maven"
}

// Detect checks if the given filesystem matches any known rules
func Detect(path fs.FS) (Rule, error) {
	// For now, only check for Spring Boot with Maven
	exists, err := fs.Stat(path, "pom.xml")
	if err == nil && !exists.IsDir() {
		return Rule{
			Name: "spring-boot",
			Tool: "maven",
		}, nil
	}

	return Rule{}, nil
}
