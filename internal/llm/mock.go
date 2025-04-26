package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MockClient is a mock implementation of the LLM client for testing
type MockClient struct{}

// Chat implements the Client interface for MockClient
func (c *MockClient) Chat(prompt string, model string) (string, error) {
	// Get the workspace root
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %v", err)
	}

	// Determine fixture path based on model
	var fixturePath string
	switch model {
	case "dockerfile":
		fixturePath = filepath.Join(wd, "testdata", "fixtures", "dockerfile", "response.json")
	case "facts":
		fixturePath = filepath.Join(wd, "testdata", "fixtures", "facts", "response.json")
	default:
		return "", fmt.Errorf("unsupported model: %s", model)
	}

	// Read fixture file
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return "", fmt.Errorf("error reading fixture file: %v", err)
	}

	// Parse fixture file
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return "", fmt.Errorf("error parsing fixture file: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in fixture file")
	}

	return response.Choices[0].Message.Content, nil
}
