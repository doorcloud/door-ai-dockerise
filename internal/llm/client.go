package llm

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/sashabaranov/go-openai"
)

// Facts represents the analyzed facts about a project
type Facts struct {
	Language  string
	Framework string
	BuildTool string
	BuildCmd  string
	StartCmd  string
	Ports     []int
	Health    string
	BaseImage string
	Env       map[string]string
	Artifact  string
	BuildDir  string
}

// Client defines the interface for LLM interactions
type Client interface {
	Chat(model, prompt string) (string, error)
}

// New returns a mock or real LLM client based on environment variables
func New() Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	mockLLM := os.Getenv("DG_MOCK_LLM") == "1"

	if apiKey == "" || mockLLM {
		return &mockClient{}
	}
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

// OpenAIClient implements the Client interface using OpenAI's API
type OpenAIClient struct {
	client *openai.Client
}

func (c *OpenAIClient) Chat(model, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai call failed: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

// AnalyzeFacts analyzes facts about a technology stack
func (c *OpenAIClient) AnalyzeFacts(ctx context.Context, facts map[string]interface{}) (map[string]interface{}, error) {
	// Implementation of AnalyzeFacts method
	return nil, nil
}

// GenerateDockerfile generates a Dockerfile based on analyzed facts
func (c *OpenAIClient) GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error) {
	// Implementation of GenerateDockerfile method
	return "", nil
}

// InferFacts analyzes snippets and returns facts about the project
func InferFacts(ctx context.Context, snippets []string) (types.Facts, error) {
	// TODO: Implement this function
	return types.Facts{}, nil
}

// GenerateDockerfile generates a Dockerfile based on the provided facts
func GenerateDockerfile(ctx context.Context, f types.Facts, currentDF string, errLog string, attempt int) (string, error) {
	// TODO: Implement this function
	return "", nil
}

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
