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
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
{{if .HasLockfile}}RUN npm ci --silent{{else}}RUN npm install{{end}}
COPY . .
RUN npm run build

# runtime
FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY package*.json ./
{{if .HasLockfile}}RUN npm ci --silent{{else}}RUN npm install{{end}}

# if the project defines a "start" script, honour it
# otherwise install serve and serve the static build
RUN if jq -e '.scripts.start' < package.json >/dev/null; then \
      echo 'using custom start script'; \
    else \
      npm install --silent --global serve; \
    fi

# Detect port from start script or use default
RUN PORT=$(jq -r '.scripts.start' package.json | grep -oE '--port[ =]?[0-9]+|-p[ =]?[0-9]+' | grep -oE '[0-9]+' || echo "3000") && \
    echo "Detected port: $PORT" && \
    echo "export PORT=$PORT" > /app/port.sh

EXPOSE $PORT
CMD ["sh","-c",". /app/port.sh && if jq -e '.scripts.start' < package.json >/dev/null; then npm run start; else serve -s build -l $PORT; fi"]

# runtime
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
EXPOSE $PORT
# install curl if missing
RUN command -v curl >/dev/null || (apk update && apk add --no-cache curl && rm -rf /var/cache/apk/*)
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:$PORT/ || exit 1`))

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