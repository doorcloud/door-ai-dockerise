package llm

import (
	"context"
	"os"
	"testing"
)

type mockClient struct{}

func (m *mockClient) AnalyzeFacts(ctx context.Context, facts map[string]interface{}) (map[string]interface{}, error) {
	return facts, nil
}

func (m *mockClient) GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error) {
	return "FROM alpine", nil
}

func TestMockClient(t *testing.T) {
	cli := &mockClient{}
	ctx := context.Background()

	facts := map[string]interface{}{
		"language":  "go",
		"framework": "none",
	}

	// Test AnalyzeFacts
	result, err := cli.AnalyzeFacts(ctx, facts)
	if err != nil {
		t.Errorf("AnalyzeFacts failed: %v", err)
	}
	if result["language"] != "go" {
		t.Errorf("expected language to be go, got %v", result["language"])
	}

	// Test GenerateDockerfile
	dockerfile, err := cli.GenerateDockerfile(ctx, facts)
	if err != nil {
		t.Errorf("GenerateDockerfile failed: %v", err)
	}
	if dockerfile != "FROM alpine" {
		t.Errorf("expected FROM alpine, got %s", dockerfile)
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
