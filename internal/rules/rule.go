package rules

// Rule defines the interface for project type detection and Dockerfile generation
type Rule interface {
	// Match returns true if the rule applies to the given facts
	Match(facts string) bool

	// GenPrompt returns the prompt for generating a first Dockerfile
	GenPrompt() string

	// FixPrompt returns the prompt for fixing a failed Dockerfile
	FixPrompt() string
}
