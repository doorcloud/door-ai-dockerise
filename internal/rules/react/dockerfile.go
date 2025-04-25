package react

import (
	"bytes"
	"io/fs"
	"text/template"

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
EXPOSE 3000
HEALTHCHECK CMD wget -qO- http://localhost:3000/ || exit 1`))

// DockerfileGenerator implements rules.RuleWithDockerfile.
type DockerfileGenerator struct{}

func (DockerfileGenerator) Name() string {
	return "react"
}

func (DockerfileGenerator) Detect(fsys fs.FS) bool {
	return Detector{}.Detect(fsys)
}

func (DockerfileGenerator) Facts(fsys fs.FS) map[string]any {
	return FactsDetector{}.Facts(fsys)
}

func (DockerfileGenerator) Dockerfile(f *types.Facts) (string, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, nil); err != nil {
		return "", err
	}
	return buf.String(), nil
} 