package llm

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
)

// mockClient implements the Client interface using local fixtures
type mockClient struct{}

func (c *mockClient) Chat(model, prompt string) (string, error) {
	// Compute hash of the prompt
	hash := fmt.Sprintf("%x", sha1.Sum([]byte(prompt)))

	// Determine which fixture file to use based on the prompt content
	var fixtureFile string
	if model == "facts" {
		fixtureFile = "testdata/mocks/facts.json"
	} else {
		fixtureFile = "testdata/mocks/dockerfile.json"
	}

	// Read and parse the fixture file
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		return "", fmt.Errorf("mock LLM: failed to read fixture file: %v", err)
	}

	var fixtures map[string]string
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return "", fmt.Errorf("mock LLM: failed to parse fixture file: %v", err)
	}

	// Look up the response
	response, exists := fixtures[hash]
	if !exists {
		return "", fmt.Errorf("mock LLM: fixture missing for %s; add it to %s", hash, fixtureFile)
	}

	return response, nil
}
