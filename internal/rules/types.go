package rules

import (
	"errors"
	"os"

	"github.com/bmatcuk/doublestar/v4"
)

// Facts represents the detected facts about a repository
type Facts struct {
	Language     string            // "java", "node", "python"…
	Framework    string            // "spring-boot", "express", "flask"…
	BuildTool    string            // "maven", "npm", "pip", …
	BuildCmd     string            // e.g. "mvn package", "npm run build"
	BuildDir     string            // directory containing build files (e.g. ".", "backend/")
	StartCmd     string            // e.g. "java -jar app.jar", "node server.js"
	Artifact     string            // glob or relative path
	Ports        []int             // e.g. [8080], [3000]
	Health       string            // URL path or CMD
	Env          map[string]string // e.g. {"NODE_ENV": "production"}
	BaseHint     string            // e.g. "eclipse-temurin:17-jdk"
	MavenVersion string            // e.g. "3.9.6"
	DevMode      bool              // whether to include development dependencies
}

// ErrNoRule is returned when no matching rule is found
var ErrNoRule = errors.New("no matching rule found")

// StackRule represents a stack detection rule
type StackRule struct {
	Name          string            // e.g. "springboot", "nodejs"
	Signatures    []string          // glob patterns that prove the stack
	ManifestGlobs []string          // files we must feed to the LLM
	CodeGlobs     []string          // optional – biggest 2 go to LLM if needed
	MainRegex     string            // marker that identifies the main class/file
	BuildHints    map[string]string // e.g. {"builder":"maven:3.9-eclipse-temurin-21"}
}

// RuleSet represents a collection of rules
type RuleSet struct {
	rules []StackRule
}

// NewRuleSet creates a new RuleSet from a slice of rules
func NewRuleSet(rules []StackRule) *RuleSet {
	return &RuleSet{rules: rules}
}

// Detect finds the first matching rule for the given repository
func (rs *RuleSet) Detect(repo string) (*StackRule, error) {
	for _, rule := range rs.rules {
		// Check each signature pattern
		for _, pattern := range rule.Signatures {
			matches, err := doublestar.Glob(os.DirFS(repo), pattern)
			if err != nil {
				continue
			}
			if len(matches) > 0 {
				return &rule, nil
			}
		}
	}
	return nil, ErrNoRule
}
