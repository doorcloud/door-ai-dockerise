package rules

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLStackRule represents a stack detection rule in YAML format
type YAMLStackRule struct {
	Kind   string `yaml:"kind"`
	ID     string `yaml:"id"`
	Detect struct {
		Globs         []string `yaml:"globs"`
		ContainsRegex string   `yaml:"containsRegex"`
	} `yaml:"detect"`
	Hints YAMLRuleHints `yaml:"hints"`
}

// YAMLRuleHints contains the hints for building and running the application in YAML format
type YAMLRuleHints struct {
	Language  string `yaml:"language"`
	Framework string `yaml:"framework"`
	Build     struct {
		Tool string `yaml:"tool"`
		Cmd  string `yaml:"cmd"`
		Dir  string `yaml:"dir"`
	} `yaml:"build"`
	Ports     []int  `yaml:"ports"`
	Health    string `yaml:"health"`
	BaseImage string `yaml:"baseImage"`
}

// YAMLRuleSet represents a collection of YAML rules
type YAMLRuleSet struct {
	rules []YAMLStackRule
}

// NewYAMLRuleSet creates a new YAMLRuleSet from a slice of rules
func NewYAMLRuleSet(rules []YAMLStackRule) *YAMLRuleSet {
	return &YAMLRuleSet{rules: rules}
}

// Detect finds the first matching rule for the given repository
func (rs *YAMLRuleSet) Detect(repo string) (*YAMLStackRule, error) {
	// TODO: Implement detection logic
	return nil, nil
}

// LoadRules loads all YAML rules from the given directory
func LoadRules(dir string) ([]YAMLStackRule, error) {
	var rules []YAMLStackRule
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, _ error) error {
		if !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		var r YAMLStackRule
		if err := decodeYAMLFile(p, &r); err != nil {
			return err
		}

		rules = append(rules, r)
		return nil
	})
	return rules, err
}

// decodeYAMLFile decodes a YAML file into the given value
func decodeYAMLFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}
