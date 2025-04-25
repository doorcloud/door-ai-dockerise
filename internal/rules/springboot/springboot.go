package springboot

import (
	"strings"

	"github.com/doorcloud/door-ai-dockerise/internal/registry"
)

type rule struct{}

func init() {
	registry.Register(rule{})
}

func (rule) Match(facts string) bool {
	return strings.Contains(strings.ToLower(facts), `"framework":"spring boot"`)
}

func (rule) GenPrompt() string {
	return `You are a Dockerfile expert. Generate a Dockerfile for a Spring Boot application.
The Dockerfile should:
1. Use eclipse-temurin:17-jdk as base image
2. Set /app as working directory
3. Copy the source code
4. Build with Maven wrapper or Gradle wrapper
5. Expose port 8080
6. Add a health check
7. Set the JAR file as entrypoint`
}

func (rule) FixPrompt() string {
	return `You are a Dockerfile expert. Fix the Dockerfile for a Spring Boot application.
The previous attempt failed with this error:
%s

Current Dockerfile:
%s

The Dockerfile should:
1. Use eclipse-temurin:17-jdk as base image
2. Set /app as working directory
3. Copy the source code
4. Build with Maven wrapper or Gradle wrapper
5. Expose port 8080
6. Add a health check
7. Set the JAR file as entrypoint`
}
