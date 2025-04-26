package mock

import (
	"context"
	"io/fs"

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
	if dockerfile, ok := m.Responses[facts.StackType]; ok {
		return dockerfile, nil
	}
	return "", nil
}
