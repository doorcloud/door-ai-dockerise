package openai

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
}

func New(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c *Client) GenerateDockerfile(ctx context.Context, facts []string) (string, error) {
	prompt := "Generate a Dockerfile for a project with the following facts:\n" + strings.Join(facts, "\n")

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
