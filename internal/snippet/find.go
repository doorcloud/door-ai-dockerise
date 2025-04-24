package snippet

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DetectBuildDir finds the directory containing the build manifest that matches the framework signature.
func DetectBuildDir(repo string) string {
	var buildDir string
	var shortestPath string
	var shortestDepth int = 999

	// First, find all POM files
	pomFiles := make(map[string]bool)
	filepath.WalkDir(repo, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Limit depth to 5 levels
			relPath, _ := filepath.Rel(repo, path)
			if strings.Count(relPath, string(filepath.Separator)) > 5 {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Base(path) == "pom.xml" {
			pomFiles[path] = true
		}
		return nil
	})

	// Then, find the shortest path containing a Spring Boot POM
	for path := range pomFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Check for Spring Boot signature
		if bytes.Contains(content, []byte("<artifactId>spring-boot-starter")) {
			dir := filepath.Dir(path)
			relPath, _ := filepath.Rel(repo, dir)
			depth := strings.Count(relPath, string(filepath.Separator))

			if depth < shortestDepth {
				shortestDepth = depth
				shortestPath = dir
			}
		}
	}

	if shortestPath != "" {
		buildDir = shortestPath
	} else {
		// Fallback to other build files if no Spring Boot POM found
		filepath.WalkDir(repo, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Limit depth to 3 levels
				relPath, _ := filepath.Rel(repo, path)
				if strings.Count(relPath, string(filepath.Separator)) > 3 {
					return filepath.SkipDir
				}
				return nil
			}

			// Check for other build files
			switch filepath.Base(path) {
			case "build.gradle", "build.gradle.kts":
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				// Check for Spring Boot signature
				if bytes.Contains(content, []byte("org.springframework.boot")) {
					buildDir = filepath.Dir(path)
					return filepath.SkipDir
				}
			case "package.json":
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				// Check for React signature
				if bytes.Contains(content, []byte("react-dom")) {
					buildDir = filepath.Dir(path)
					return filepath.SkipDir
				}
			case "requirements.txt":
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				// Check for Flask signature
				if bytes.Contains(content, []byte("flask")) {
					buildDir = filepath.Dir(path)
					return filepath.SkipDir
				}
			}
			return nil
		})
	}

	return buildDir
}
