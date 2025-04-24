package rule

import (
	"os"
	"path/filepath"
	"strings"
)

// Project represents a project directory with its root path.
type Project struct {
	Root string
}

// NewProject creates a new project instance.
func NewProject(path string) (*Project, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &Project{Root: absPath}, nil
}

// FindProjectRoot searches for the project root directory.
func FindProjectRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// Start from the given path and move up until we find a project root
	for path := absPath; path != "/"; path = filepath.Dir(path) {
		// Check for common project root indicators
		if isProjectRoot(path) {
			return path, nil
		}
	}

	// If we reach here, we didn't find a project root
	return absPath, nil
}

// isProjectRoot checks if the given path is a project root.
func isProjectRoot(path string) bool {
	// Check for common project root files
	indicators := []string{
		"package.json",     // Node.js
		"pom.xml",          // Java/Maven
		"build.gradle",     // Java/Gradle
		"requirements.txt", // Python
		"go.mod",           // Go
		".git",             // Git repository
	}

	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(path, indicator)); err == nil {
			return true
		}
	}

	return false
}

// Join returns the absolute path joined with the project root.
func (p *Project) Join(elem ...string) string {
	return filepath.Join(append([]string{p.Root}, elem...)...)
}

// Relative returns the path relative to the project root.
func (p *Project) Relative(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(absPath, p.Root), nil
}

// Exists checks if a file or directory exists in the project.
func (p *Project) Exists(elem ...string) bool {
	path := p.Join(elem...)
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory in the project.
func (p *Project) IsDir(elem ...string) bool {
	path := p.Join(elem...)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IsFile checks if a path is a file in the project.
func (p *Project) IsFile(elem ...string) bool {
	path := p.Join(elem...)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
