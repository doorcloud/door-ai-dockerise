package llm

import (
	"os"
	"testing"
)

func TestMockClient(t *testing.T) {
	cli := &mockClient{}

	resp, err := cli.Chat("facts", `{"language": "go", "framework": "none"}`)
	if err != nil {
		t.Errorf("Chat() failed: %v", err)
	}
	if !isValidJSON(resp) {
		t.Errorf("Chat() response is not valid JSON: %v", resp)
	}
}

func TestNew_ReturnsMockClientWhenNoAPIKey(t *testing.T) {
	// Save and restore environment
	originalAPIKey := os.Getenv("OPENAI_API_KEY")
	originalMockLLM := os.Getenv("DG_MOCK_LLM")
	defer func() {
		os.Setenv("OPENAI_API_KEY", originalAPIKey)
		os.Setenv("DG_MOCK_LLM", originalMockLLM)
	}()

	// Test cases
	tests := []struct {
		name     string
		apiKey   string
		mockLLM  string
		wantMock bool
	}{
		{"No API Key", "", "", true},
		{"Empty API Key", "", "0", true},
		{"With API Key", "test-key", "", false},
		{"Force Mock", "test-key", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OPENAI_API_KEY", tt.apiKey)
			os.Setenv("DG_MOCK_LLM", tt.mockLLM)

			client := New()
			_, isMock := client.(*mockClient)
			if isMock != tt.wantMock {
				t.Errorf("New() = %T, want mock=%v", client, tt.wantMock)
			}
		})
	}
}

func TestMockClient_Chat(t *testing.T) {
	client := &mockClient{}

	tests := []struct {
		name    string
		model   string
		prompt  string
		wantErr bool
	}{
		{
			name:    "Facts Model",
			model:   "facts",
			prompt:  "test prompt",
			wantErr: false,
		},
		{
			name:    "Dockerfile Model",
			model:   "dockerfile",
			prompt:  "test prompt",
			wantErr: false,
		},
		{
			name:    "Unknown Model",
			model:   "unknown",
			prompt:  "test prompt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Chat(tt.model, tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.model == "facts" && !isValidJSON(resp) {
				t.Errorf("Chat() response is not valid JSON: %v", resp)
			}
		})
	}
}

func isValidJSON(s string) bool {
	return len(s) > 0 && s[0] == '{' && s[len(s)-1] == '}'
}
