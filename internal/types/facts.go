package types

// Facts represents the detected facts about a technology stack
type Facts struct {
	Language    string            `json:"language"`     // "java", "node", "python"…
	Framework   string            `json:"framework"`    // "spring-boot", "express", "flask"…
	BuildTool   string            `json:"build_tool"`   // "maven", "npm", "pip", …
	BuildCmd    string            `json:"build_cmd"`    // e.g. "mvn package", "npm run build"
	BuildDir    string            `json:"build_dir"`    // directory containing build files
	StartCmd    string            `json:"start_cmd"`    // e.g. "java -jar app.jar"
	Artifact    string            `json:"artifact"`     // glob or relative path
	Ports       []int             `json:"ports"`        // e.g. [8080], [3000]
	Health      string            `json:"health"`       // URL path or CMD
	Env         map[string]string `json:"env"`          // e.g. {"NODE_ENV": "production"}
	BaseImage   string            `json:"base_image"`   // e.g. "eclipse-temurin:17-jdk"
	HasLockfile bool              `json:"has_lockfile"` // whether package-lock.json exists
}
