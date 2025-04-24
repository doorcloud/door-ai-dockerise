package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	// BuildTimeout is the maximum time to wait for a Docker build
	BuildTimeout time.Duration
	// Debug enables debug logging
	Debug bool
	// MvnVersion is the Maven version to use
	MvnVersion string
}

// New creates a new configuration with default values
func New() *Config {
	return &Config{
		BuildTimeout: 15 * time.Minute,
		Debug:        false,
		MvnVersion:   "3.8.4",
	}
}
