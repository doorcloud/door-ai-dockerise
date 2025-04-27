package openai

import (
	"context"
	"fmt"
	"io/fs"

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

// GatherFacts implements the core.ChatCompletion interface
func (o *OpenAI) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	// First, gather basic facts from the stack info
	facts := core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
		Port:      stack.Port,
	}

	// Create a prompt to analyze the project
	prompt := fmt.Sprintf(`Analyze this %s project and provide the following information in JSON format:
1. Required dependencies and their versions
2. Build commands and configuration
3. Runtime requirements
4. Environment variables needed
5. Port number if not already specified
6. Any special considerations for containerization

Project type: %s
Build tool: %s
Detected files: %v

Please provide the information in a structured JSON format.`,
		stack.Name, stack.Name, stack.BuildTool, stack.DetectedFiles)

	// Call OpenAI API to analyze the project
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a Docker expert. Analyze the project and provide detailed information for containerization.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return facts, fmt.Errorf("failed to analyze project: %w", err)
	}

	// Log the analysis for debugging purposes
	if len(resp.Choices) > 0 {
		fmt.Printf("Project analysis: %s\n", resp.Choices[0].Message.Content)
	}

	return facts, nil
}

// GenerateDockerfile implements the core.ChatCompletion interface
func (o *OpenAI) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	prompt := fmt.Sprintf("Generate a Dockerfile for a %s project using %s as the build tool.", facts.StackType, facts.BuildTool)

	resp, err := o.client.CreateChatCompletion(
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

// Complete implements the core.ChatCompletion interface
func (o *OpenAI) Complete(ctx context.Context, messages []core.Message) (string, error) {
	// Convert core.Message to openai.ChatCompletionMessage
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call OpenAI API
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4,
			Messages: openaiMessages,
		},
	)
	if err != nil {
		return "", err
	}

	// Return the first choice's content
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}
