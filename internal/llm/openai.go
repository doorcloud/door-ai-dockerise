package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// openAIClient implements the Client interface using OpenAI's API
type openAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string) Client {
	return &openAIClient{
		client: openai.NewClient(apiKey),
	}
}

// AnalyzeFacts analyzes facts about a technology stack using OpenAI
func (c *openAIClient) AnalyzeFacts(ctx context.Context, facts map[string]interface{}) (map[string]interface{}, error) {
	// Convert facts to JSON for the prompt
	factsJSON, err := json.Marshal(facts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal facts: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze the following technology stack facts and provide additional insights:
%s

Provide the response in JSON format with any additional facts or recommendations.`, factsJSON)

	// TODO: Make API call to OpenAI using the prompt
	// For now, just return the input facts
	_ = prompt // Use the prompt variable to avoid linter error
	return facts, nil
}

// GenerateDockerfile generates a Dockerfile based on analyzed facts
func (c *openAIClient) GenerateDockerfile(ctx context.Context, facts map[string]interface{}) (string, error) {
	// Convert facts to JSON for the prompt
	factsJSON, err := json.Marshal(facts)
	if err != nil {
		return "", fmt.Errorf("failed to marshal facts: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a Dockerfile based on the following technology stack facts:
%s

Provide only the Dockerfile content, no explanations or markdown formatting.`, factsJSON)

	// TODO: Make API call to OpenAI using the prompt
	// For now, return a basic Dockerfile
	_ = prompt // Use the prompt variable to avoid linter error
	return `FROM alpine:latest
WORKDIR /app
COPY . .
CMD ["echo", "Hello, World!"]`, nil
}

// Chat sends a prompt to OpenAI and returns the response
func (c *openAIClient) Chat(ctx context.Context, prompt string) (string, error) {
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
