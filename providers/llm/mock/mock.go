package mock

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type MockClient struct {
	responses map[string]string
}

func NewMockClient() *MockClient {
	return &MockClient{
		responses: make(map[string]string),
	}
}

func (m *MockClient) SetResponse(prompt string, response string) {
	m.responses[prompt] = response
}

func (m *MockClient) Complete(ctx context.Context, messages []core.Message) (string, error) {
	// For testing, just return the last message content
	if len(messages) > 0 {
		content := messages[len(messages)-1].Content
		// Check if this is a React prompt
		if strings.Contains(content, "framework: react") {
			return `FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./
RUN npm install --production

EXPOSE 3001
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/ || exit 1

CMD ["npm", "start"]`, nil
		}
		return content, nil
	}
	return "FROM ubuntu:latest\n", nil
}

func (m *MockClient) Generate(ctx context.Context, facts core.Facts) (string, error) {
	key := facts.StackType + ":" + facts.BuildTool
	if response, ok := m.responses[key]; ok {
		return response, nil
	}
	return "FROM ubuntu:latest\n", nil
}

func (m *MockClient) Fix(ctx context.Context, prevDockerfile string, buildErr string) (string, error) {
	// If the error contains "ERROR", return a fixed version
	if strings.Contains(buildErr, "ERROR") {
		return "FROM ubuntu:latest\nRUN apt-get update\n", nil
	}
	return prevDockerfile, nil
}

func (m *MockClient) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}, nil
}

func (m *MockClient) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	return m.Generate(ctx, facts)
}
