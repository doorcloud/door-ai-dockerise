package generate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// StaticGenerator implements the ChatCompletion interface with static templates
type StaticGenerator struct{}

// New creates a new StaticGenerator
func New() *StaticGenerator {
	return &StaticGenerator{}
}

// GatherFacts implements the ChatCompletion interface
func (g *StaticGenerator) GatherFacts(ctx context.Context, fsys fs.FS, stack core.StackInfo) (core.Facts, error) {
	return core.Facts{
		StackType: stack.Name,
		BuildTool: stack.BuildTool,
	}, nil
}

// GenerateDockerfile implements the ChatCompletion interface
func (g *StaticGenerator) GenerateDockerfile(ctx context.Context, facts core.Facts) (string, error) {
	switch facts.StackType {
	case "react":
		return `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`, nil
	case "node":
		return `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE 3000
CMD ["npm", "start"]`, nil
	case "springboot":
		return `FROM maven:3.8-openjdk-11 AS build
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

FROM openjdk:11-jre-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]`, nil
	default:
		return "", fmt.Errorf("unsupported stack: %s", facts.StackType)
	}
}
