package rules

import "errors"

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
