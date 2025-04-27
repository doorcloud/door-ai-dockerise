package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	LLMModel       string
	LLMTemperature float64
	MaxAttempts    int
	DockerTimeout  time.Duration
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Parse LLM model
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gpt-4-mini" // default model
	}

	// Parse temperature
	tempStr := os.Getenv("LLM_TEMPERATURE")
	temperature := 0.7 // default temperature
	if tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			temperature = temp
		}
	}

	// Parse max attempts
	maxAttemptsStr := os.Getenv("MAX_ATTEMPTS")
	maxAttempts := 3 // default max attempts
	if maxAttemptsStr != "" {
		if attempts, err := strconv.Atoi(maxAttemptsStr); err == nil {
			maxAttempts = attempts
		}
	}

	// Parse docker timeout
	timeoutStr := os.Getenv("DOCKER_TIMEOUT")
	timeout := 5 * time.Minute // default timeout
	if timeoutStr != "" {
		if duration, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = duration
		}
	}

	return &Config{
		LLMModel:       model,
		LLMTemperature: temperature,
		MaxAttempts:    maxAttempts,
		DockerTimeout:  timeout,
	}, nil
}
