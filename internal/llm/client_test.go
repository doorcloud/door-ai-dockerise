package llm

import (
	"context"
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
