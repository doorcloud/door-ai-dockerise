package react

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"github.com/doorcloud/door-ai-dockerise/internal/llm"
	"github.com/doorcloud/door-ai-dockerise/internal/types"
)

var tpl = template.Must(template.New("reactDockerfile").Parse(`# build stage
FROM node:18-alpine AS build
WORKDIR /app
COPY . .
RUN npm ci && npm run build

# runtime
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE {{index .Ports 0}}
HEALTHCHECK CMD wget -qO- http://localhost:{{index .Ports 0}}/ || exit 1`))

// DockerfileGenerator implements rules.RuleWithDockerfile.
type DockerfileGenerator struct{}

func (DockerfileGenerator) Name() string {
	return "react"
}

func (DockerfileGenerator) Detect(fsys fs.FS) bool {
	return (&ReactDetector{}).Detect(fsys)
}

func (DockerfileGenerator) Facts(fsys fs.FS) map[string]any {
	return (&ReactDetector{}).Facts(fsys)
}

func (DockerfileGenerator) Dockerfile(f *types.Facts) (string, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, f); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (DockerfileGenerator) GenerateDockerfile(ctx context.Context, cli llm.Client, facts map[string]any, prevDockerfile string, attempt int) (string, error) {
	var prompt string
	if attempt == 0 {
		prompt = fmt.Sprintf(`Create a multi-stage Dockerfile for a React static site with these facts:
- build_tool: %s
- build_cmd: %s
- ports: %v
- health: %s

The Dockerfile should:
1. Use a Node.js base image for building
2. Copy the source code and run the build command
3. Use a lightweight web server (nginx) for serving the static files
4. Expose the correct port
5. Include a health check`, 
			facts["build_tool"],
			facts["build_cmd"],
			facts["ports"],
			facts["health"])
	} else {
		// On retry, include the previous Dockerfile and error log
		prompt = fmt.Sprintf(`Fix this Dockerfile that failed to build:

Previous Dockerfile:
%s

Error log:
%s

Please fix only what's necessary while keeping the multi-stage structure.`, 
			prevDockerfile,
			strings.Join(strings.Split(prevDockerfile, "\n")[:80], "\n"))
	}

	dockerfile, err := cli.Chat(prompt, "dockerfile")
	if err != nil {
		return "", fmt.Errorf("generate dockerfile: %w", err)
	}

	return dockerfile, nil
} 