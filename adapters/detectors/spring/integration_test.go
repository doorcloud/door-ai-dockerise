package spring_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
)

func TestDetectSpringBootRepos(t *testing.T) {
	t.Logf("Starting Spring Boot repository detection test")

	// Get absolute path to testdata directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	testdataRoot := filepath.Join(filepath.Dir(wd), "..", "..", "testdata", "e2e")

	// Map of test directories to their actual project paths (relative to the repo root)
	projectPaths := map[string]string{
		"gradle_groovy_simple": "SmallestSpringApp",
		"gradle_kts_multi":     "",
		"kotlin_gradle_demo":   "",
		"maven_multimodule":    "",
		"plain_java":           "",
	}

	walk := func(base string, want bool) {
		fullPath := filepath.Join(testdataRoot, base)
		t.Logf("Walking directory: %s (expecting Spring Boot: %v)", fullPath, want)
		_ = filepath.WalkDir(fullPath, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				t.Logf("Error walking path %s: %v", p, err)
				return nil
			}
			if !d.IsDir() || p == fullPath {
				return nil
			}

			dirName := filepath.Base(p)

			// Skip if this is not a project root we want to test
			if projectPath, ok := projectPaths[dirName]; ok {
				testPath := p
				if projectPath != "" {
					testPath = filepath.Join(p, projectPath)
				}

				got := spring.IsSpringBoot(testPath)
				t.Logf("Checking %s: got=%v want=%v", testPath, got, want)
				if got != want {
					t.Errorf("%s → got %v want %v", testPath, got, want)
				}
			}

			return filepath.SkipDir // one repo per dir
		})
	}

	walk("spring_positive", true)
	walk("non_spring", false)
}
