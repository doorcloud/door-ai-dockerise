package mock

import (
	"context"

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

// Chat implements the ChatCompletion interface
func (m *MockLLM) Chat(ctx context.Context, msgs []core.Message) (core.Message, error) {
	// Extract the stack type from the last message
	lastMsg := msgs[len(msgs)-1].Content
	var stackType string
	if lastMsg == "react" || lastMsg == "springboot" || lastMsg == "node" {
		stackType = lastMsg
	} else {
		stackType = "node" // default
	}

	return core.Message{
		Role:    "assistant",
		Content: m.Responses[stackType],
	}, nil
}
