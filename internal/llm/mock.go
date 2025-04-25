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
	// Get the fixture path based on the model
	fixturePath := filepath.Join("testdata", "fixtures", model, "response.json")

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
