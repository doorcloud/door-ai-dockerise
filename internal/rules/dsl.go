package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// RuleConfig represents the JSON configuration for a technology stack rule.
type RuleConfig struct {
	Name          string            `json:"name"`           // e.g. "springboot", "nodejs"
	Language      string            `json:"language"`       // e.g. "java", "javascript"
	Framework     string            `json:"framework"`      // e.g. "spring-boot", "vue"
	Signatures    []string          `json:"signatures"`     // glob patterns that prove the stack
	ManifestGlobs []string          `json:"manifest_globs"` // files we must feed to the LLM
	CodeGlobs     []string          `json:"code_globs"`     // optional – biggest 2 go to LLM if needed
	MainRegex     string            `json:"main_regex"`     // marker that identifies the main class/file
	BuildHints    map[string]string `json:"build_hints"`    // e.g. {"builder":"maven:3.9-eclipse-temurin-21"}
	PortConfigs   []PortConfig      `json:"port_configs"`   // config files to check for ports
}

// PortConfig defines where to look for port configurations
type PortConfig struct {
	FilePattern string `json:"file_pattern"` // e.g. "**/application*.{yml,yaml,properties}"
	KeyPattern  string `json:"key_pattern"`  // e.g. "server.port"
}

// LoadRuleConfigs loads rule configurations from a directory
func LoadRuleConfigs(configDir string) (map[string]*RuleConfig, error) {
	configs := make(map[string]*RuleConfig)

	// Find all JSON files in the config directory
	matches, err := doublestar.Glob(os.DirFS(configDir), "**/*.json")
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		path := filepath.Join(configDir, match)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var config RuleConfig
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, err
		}

		configs[config.Name] = &config
	}

	return configs, nil
}

// FindBestManifest finds the most appropriate manifest file in the repository
func FindBestManifest(repo string, config *RuleConfig) (string, error) {
	var bestMatch string
	var bestScore int

	for _, pattern := range config.ManifestGlobs {
		matches, err := doublestar.Glob(os.DirFS(repo), pattern)
		if err != nil {
			return "", err
		}

		for _, match := range matches {
			score := scoreManifest(match, config)
			if score > bestScore {
				bestScore = score
				bestMatch = match
			}
		}
	}

	return bestMatch, nil
}

// scoreManifest assigns a score to a manifest file based on its location and name
func scoreManifest(path string, config *RuleConfig) int {
	score := 0

	// Prefer manifests in root directory
	if !strings.Contains(path, "/") {
		score += 10
	}

	// Prefer standard manifest names
	base := filepath.Base(path)
	switch {
	case base == "pom.xml" && config.BuildHints["builder"] == "maven":
		score += 5
	case base == "package.json" && config.BuildHints["builder"] == "npm":
		score += 5
	case base == "go.mod" && config.BuildHints["builder"] == "go":
		score += 5
	}

	return score
}
