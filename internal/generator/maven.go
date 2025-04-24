package generator

// DefaultMavenBuildCmd returns the default Maven build command
func DefaultMavenBuildCmd() string {
	return "./mvnw -B -ntp package -DskipTests"
}
