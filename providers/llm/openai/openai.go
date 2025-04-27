package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type Provider struct {
	apiKey string
	client *http.Client
	model  string
}

func NewProvider(apiKey string) *Provider {
	// If API key is empty, try to get it from environment
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	// Get model from environment, default to gpt-3.5-turbo
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &Provider{
		apiKey: apiKey,
		client: &http.Client{},
		model:  model,
	}
}

func (p *Provider) Complete(ctx context.Context, messages []core.Message) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not provided")
	}

	payload := struct {
		Model    string         `json:"model"`
		Messages []core.Message `json:"messages"`
	}{
		Model:    p.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned status code %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no completion returned")
	}

	return result.Choices[0].Message.Content, nil
}

func (p *Provider) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	// TODO: Implement fact gathering using LLM
	return core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}, nil
}

func (p *Provider) Generate(ctx context.Context, facts core.Facts) (string, error) {
	messages := []core.Message{
		{
			Role:    "system",
			Content: "You are a Dockerfile expert. Generate a Dockerfile based on the provided facts.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Generate a Dockerfile for a %s application using %s as the build tool.", facts.StackType, facts.BuildTool),
		},
	}

	return p.Complete(ctx, messages)
}

func (p *Provider) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	messages := []core.Message{
		{
			Role:    "system",
			Content: "You are a Dockerfile expert. Fix the Dockerfile based on the build error.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Fix this Dockerfile:\n%s\n\nBuild error:\n%s", prevDockerfile, buildErr),
		},
	}

	return p.Complete(ctx, messages)
}
