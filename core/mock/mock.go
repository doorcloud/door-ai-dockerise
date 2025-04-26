package mock

import (
	"context"
	"io/fs"
	"strings"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// MockLLM implements ChatCompletion with canned responses
type MockLLM struct {
	Responses map[string]string
}

// NewMockLLM creates a new MockLLM with default canned responses
func NewMockLLM() *MockLLM {
	return &MockLLM{
		Responses: map[string]string{
			"react": `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN if [ -f package-lock.json ]; then \
        npm ci --silent ; \
    else \
        npm install --production --silent ; \
    fi
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`,
			"springboot": `FROM maven:3.8-openjdk-17 AS build
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

FROM openjdk:17-jdk-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]`,
			"node": `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN if [ -f package-lock.json ]; then \
        npm ci --silent ; \
    else \
        npm install --production --silent ; \
    fi
COPY . .
EXPOSE 3000
CMD ["npm", "start"]`,
		},
	}
}

// GatherFacts implements the ChatCompletion interface
func (m *MockLLM) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}, nil
}

// GenerateDockerfile implements the ChatCompletion interface
func (m *MockLLM) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	key := facts.StackType
	if response, ok := m.Responses[key]; ok {
		return response, nil
	}
	return "FROM ubuntu:latest\n", nil
}

func (m *MockLLM) Complete(ctx context.Context, messages []core.Message) (string, error) {
	// Extract stack type from messages
	for _, msg := range messages {
		if msg.Role == "user" && strings.Contains(msg.Content, "Stack type:") {
			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Stack type:") {
					stackType := strings.TrimSpace(strings.TrimPrefix(line, "Stack type:"))
					if response, ok := m.Responses[stackType]; ok {
						return response, nil
					}
				}
			}
		}
	}
	return "FROM ubuntu:latest\n", nil
}
