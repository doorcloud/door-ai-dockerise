package config

import (
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration settings for the application
type Config struct {
	// OpenAI settings
	OpenAIKey      string  `envconfig:"OPENAI_API_KEY"`
	OpenAIModel    string  `envconfig:"OPENAI_MODEL" default:"gpt-4"`
	OpenAITemp     float64 `envconfig:"OPENAI_TEMPERATURE" default:"0.7"`
	OpenAILogLevel string  `envconfig:"OPENAI_LOG_LEVEL" default:"info"`

	// Application settings
	Debug        bool          `envconfig:"DG_DEBUG" default:"false"`
	E2E          bool          `envconfig:"DG_E2E" default:"false"`
	BuildTimeout time.Duration `envconfig:"DG_BUILD_TIMEOUT" default:"15m"`
	MvnVersion   string        `envconfig:"DG_MVN_VERSION" default:"3.9.6"`
}

// Load loads the configuration from environment variables
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// GetM2Cache returns the Maven cache directory path
func (c *Config) GetM2Cache() string {
	return os.Getenv("HOME") + "/.m2"
}
