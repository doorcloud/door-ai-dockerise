package llm

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Generate calls the ChatCompletion API with the given prompt.
func Generate(prompt, apiKey string) (string, error) {
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     "gpt-4o-mini",
			MaxTokens: 500,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: prompt},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
