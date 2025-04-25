package llm

import (
	"os"
	"testing"
)

func TestMockClient(t *testing.T) {
	cli := &MockClient{}

	resp, err := cli.Chat(`{"language": "go", "framework": "none"}`, "facts")
	if err != nil {
		t.Errorf("Chat() failed: %v", err)
	}
	if !isValidJSON(resp) {
		t.Errorf("Chat() response is not valid JSON: %v", resp)
	}
}

func TestNew_ReturnsMockClientWhenNoAPIKey(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "")
	os.Setenv("DG_MOCK_LLM", "1")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("DG_MOCK_LLM")

	client := New()
	if _, ok := client.(*MockClient); !ok {
		t.Error("Expected mock client when no API key and DG_MOCK_LLM=1")
	}
}

func TestNew_ReturnsRealClientWhenAPIKey(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("DG_MOCK_LLM", "0")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("DG_MOCK_LLM")

	client := New()
	if _, ok := client.(*MockClient); ok {
		t.Error("Expected real client when API key is set")
	}
}

func TestNew_ReturnsMockClientWhenMockEnvSet(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("DG_MOCK_LLM", "1")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("DG_MOCK_LLM")

	client := New()
	if _, ok := client.(*MockClient); !ok {
		t.Error("Expected mock client when DG_MOCK_LLM=1")
	}
}

func TestMockClient_Chat(t *testing.T) {
	client := &MockClient{}

	tests := []struct {
		name    string
		prompt  string
		model   string
		wantErr bool
	}{
		{
			name:    "Facts Model",
			prompt:  "test prompt",
			model:   "facts",
			wantErr: false,
		},
		{
			name:    "Dockerfile Model",
			prompt:  "test prompt",
			model:   "dockerfile",
			wantErr: false,
		},
		{
			name:    "Unknown Model",
			prompt:  "test prompt",
			model:   "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Chat(tt.prompt, tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.model == "facts" && !isValidJSON(resp) {
				t.Errorf("Chat() response is not valid JSON: %v", resp)
			}
		})
	}
}

func isValidJSON(s string) bool {
	return len(s) > 0 && s[0] == '{' && s[len(s)-1] == '}'
}
