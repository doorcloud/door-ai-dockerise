package llm

import (
	"context"
	"fmt"

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
	Chat(ctx context.Context, prompt string) (string, error)
}

// OpenAIClient implements Client using OpenAI's API
type OpenAIClient struct {
	client *openai.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) Client {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

// Chat sends a prompt to OpenAI and returns the response
func (c *OpenAIClient) Chat(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
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
