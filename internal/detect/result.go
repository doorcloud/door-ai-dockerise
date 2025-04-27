package detect

// Result represents the output of a detection process
type Result struct {
	// Stack information
	StackName    string `json:"stack_name"`
	StackVersion string `json:"stack_version"`

	// Build information
	BuildTool    string `json:"build_tool"`
	BuildCommand string `json:"build_command"`
	Artifact     string `json:"artifact"`

	// Runtime information
	Ports       []int             `json:"ports"`
	HealthCheck string            `json:"health_check"`
	Environment map[string]string `json:"environment"`

	// Dependencies
	Dependencies []string `json:"dependencies"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// NewResult creates a new Result struct with default values
func NewResult() Result {
	return Result{
		Environment: make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
}
