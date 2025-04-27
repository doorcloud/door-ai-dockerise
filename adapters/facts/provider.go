package facts

// Provider defines the interface for gathering project facts
type Provider interface {
	// Gather collects facts about the project
	Gather(projectDir string) (map[string]interface{}, error)
}
