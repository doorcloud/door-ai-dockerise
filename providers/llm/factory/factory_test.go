package factory

import (
	"os"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Set up OpenAI API key for testing
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "openai provider",
			provider: "openai",
			wantErr:  false,
		},
		{
			name:     "ollama provider",
			provider: "ollama",
			wantErr:  false,
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.provider)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Implements(t, (*core.ChatCompletion)(nil), got)
			}
		})
	}
}
