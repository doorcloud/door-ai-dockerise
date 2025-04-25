package rules

// Rule defines the interface for project type detection and Dockerfile generation
type Rule interface {
	Name() string
	Detect(repo string) bool          // true if rule matches repo
	Facts(repo string) map[string]any // can return nil for now
}
