package generate

import (
	"context"
	"os"

	"github.com/doorcloud/door-ai-dockerise/core"
)

type Static struct{}

func NewStatic() *Static {
	return &Static{}
}

func (s *Static) Generate(ctx context.Context, stack core.StackInfo, facts []core.Fact) (string, error) {
	// Create a temporary directory for the filesystem
	dir, err := os.MkdirTemp("", "generate-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	// Generate Dockerfile based on the stack type
	switch stack.Name {
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
		return `FROM maven:3.8-openjdk-17 AS build
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY src ./src
RUN mvn package -DskipTests

FROM openjdk:17-jdk-slim
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8080
CMD ["java", "-jar", "app.jar"]`, nil
	default:
		return "", nil
	}
}
