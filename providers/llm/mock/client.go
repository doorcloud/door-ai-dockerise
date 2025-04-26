package mock

import (
	"context"
	"strings"
)

type Client struct {
	responses map[string]string
}

func New(responses map[string]string) *Client {
	return &Client{
		responses: responses,
	}
}

func (c *Client) GenerateDockerfile(ctx context.Context, facts []string) (string, error) {
	key := strings.Join(facts, "\n")
	if response, ok := c.responses[key]; ok {
		return response, nil
	}
	return "FROM node:18-alpine\nWORKDIR /app\nCOPY . .\nRUN npm install\nEXPOSE 3000\nCMD [\"npm\", \"start\"]", nil
}
