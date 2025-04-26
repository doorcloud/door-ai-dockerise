package openai

import (
	"context"
	"fmt"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/sashabaranov/go-openai"
)

// OpenAI implements core.ChatCompletion using the OpenAI API
type OpenAI struct {
	client *openai.Client
}

// New creates a new OpenAI client
func New(apiKey string) *OpenAI {
	return &OpenAI{
		client: openai.NewClient(apiKey),
	}
}

// Chat implements the core.ChatCompletion interface
func (o *OpenAI) Chat(ctx context.Context, msgs []core.Message) (core.Message, error) {
	// Convert core.Message to openai.ChatCompletionMessage
	openaiMsgs := make([]openai.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		openaiMsgs[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call OpenAI API
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: openaiMsgs,
		},
	)
	if err != nil {
		return core.Message{}, fmt.Errorf("OpenAI API call failed: %v", err)
	}

	// Convert response to core.Message
	return core.Message{
		Role:    "assistant",
		Content: resp.Choices[0].Message.Content,
	}, nil
}
