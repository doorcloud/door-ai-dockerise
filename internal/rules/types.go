package rules

import "errors"

// Facts represents the detected facts about a repository
type Facts struct {
	Language  string // "java", "node", "python"…
	Framework string // "spring-boot", "express", "flask"…
	BuildTool string // "maven", "npm", "pip", …
	BuildCmd  string
	Artifact  string // glob or relative path
	Ports     []int
	Health    string // URL path or CMD
	Env       []string
	BaseHint  string // e.g. "eclipse-temurin:17-jdk"
}

// ErrNoRule is returned when no matching rule is found
var ErrNoRule = errors.New("no matching rule found")
