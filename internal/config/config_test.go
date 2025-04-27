package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	for _, key := range []string{"LLM_MODEL", "LLM_TEMPERATURE", "MAX_ATTEMPTS", "DOCKER_TIMEOUT"} {
		if value, exists := os.LookupEnv(key); exists {
			originalEnv[key] = value
			defer os.Setenv(key, value)
		} else {
			defer os.Unsetenv(key)
		}
	}

	// Test default values
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "gpt-4-mini", cfg.LLMModel)
	require.Equal(t, 0.7, cfg.LLMTemperature)
	require.Equal(t, 3, cfg.MaxAttempts)
	require.Equal(t, 5*time.Minute, cfg.DockerTimeout)

	// Test environment overrides
	os.Setenv("LLM_MODEL", "test-model")
	os.Setenv("LLM_TEMPERATURE", "0.5")
	os.Setenv("MAX_ATTEMPTS", "5")
	os.Setenv("DOCKER_TIMEOUT", "10m")

	cfg, err = Load()
	require.NoError(t, err)
	require.Equal(t, "test-model", cfg.LLMModel)
	require.Equal(t, 0.5, cfg.LLMTemperature)
	require.Equal(t, 5, cfg.MaxAttempts)
	require.Equal(t, 10*time.Minute, cfg.DockerTimeout)

	// Test invalid values
	os.Setenv("LLM_TEMPERATURE", "invalid")
	os.Setenv("MAX_ATTEMPTS", "invalid")
	os.Setenv("DOCKER_TIMEOUT", "invalid")

	cfg, err = Load()
	require.NoError(t, err)
	require.Equal(t, 0.7, cfg.LLMTemperature)          // should fall back to default
	require.Equal(t, 3, cfg.MaxAttempts)               // should fall back to default
	require.Equal(t, 5*time.Minute, cfg.DockerTimeout) // should fall back to default
}
