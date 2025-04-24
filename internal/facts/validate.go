package facts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateBasic checks if the facts are valid and consistent.
func (f Facts) ValidateBasic() error {
	// Check required fields
	if f.BuildCmd == "" {
		return errors.New("missing build_cmd")
	}
	if f.BuildDir == "" {
		return errors.New("missing build_dir")
	}
	if f.StartCmd == "" {
		return errors.New("missing start_cmd")
	}

	// Check build tool specific requirements
	if strings.Contains(f.BuildCmd, "mvn") {
		if !strings.Contains(f.BuildCmd, "-f") && !fileExists(filepath.Join(f.BuildDir, "pom.xml")) {
			return errors.New("build cmd uses Maven but no pom.xml found in build directory")
		}
	}

	if strings.Contains(f.BuildCmd, "npm") || strings.Contains(f.BuildCmd, "yarn") {
		if !fileExists(filepath.Join(f.BuildDir, "package.json")) {
			return errors.New("build cmd uses npm/yarn but no package.json found in build directory")
		}
	}

	if strings.Contains(f.BuildCmd, "pip") || strings.Contains(f.BuildCmd, "poetry") {
		if !fileExists(filepath.Join(f.BuildDir, "requirements.txt")) && !fileExists(filepath.Join(f.BuildDir, "pyproject.toml")) {
			return errors.New("build cmd uses pip/poetry but no requirements.txt or pyproject.toml found in build directory")
		}
	}

	if strings.Contains(f.BuildCmd, "go build") {
		if !fileExists(filepath.Join(f.BuildDir, "go.mod")) {
			return errors.New("build cmd uses Go but no go.mod found in build directory")
		}
	}

	// Check artifact path
	if f.Artifact == "" {
		return errors.New("missing artifact path")
	}

	// Check ports
	if len(f.Ports) == 0 {
		return errors.New("no ports specified")
	}

	// Check environment variables
	if len(f.Env) == 0 {
		return errors.New("no environment variables specified")
	}

	return nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CheckArtifactExists verifies that the artifact exists in the build directory.
func CheckArtifactExists(facts *Facts) error {
	if facts.BuildDir == "" {
		return fmt.Errorf("build directory not set")
	}

	artifactPath := filepath.Join(facts.BuildDir, facts.Artifact)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		return fmt.Errorf("artifact %s does not exist in build directory %s", facts.Artifact, facts.BuildDir)
	}

	return nil
}

// ValidateBuildDir checks if the build directory exists and contains the artifact.
func ValidateBuildDir(f Facts, buildDir string) error {
	// First validate using the base Validate method
	if err := f.ValidateBasic(); err != nil {
		return err
	}

	// Check if artifact exists in build directory
	if !filepath.IsAbs(f.Artifact) {
		f.Artifact = filepath.Join(buildDir, f.Artifact)
	}
	if _, err := os.Stat(f.Artifact); err != nil {
		return fmt.Errorf("artifact not found: %w", err)
	}

	return nil
}
