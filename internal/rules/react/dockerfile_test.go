package react

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestDockerfileGenerator(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		expected string
	}{
		{
			name:    "react-build",
			fixture: "testdata/react-build",
			expected: `# build stage
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# runtime
FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY package*.json ./
RUN npm install

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
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:$PORT/ || exit 1`,
		},
		{
			name:    "react-start",
			fixture: "testdata/react-start",
			expected: `# build stage
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# runtime
FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY package*.json ./
RUN npm install

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
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:$PORT/ || exit 1`,
		},
		{
			name:    "react-start-port",
			fixture: "testdata/react-start-port",
			expected: `# build stage
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# runtime
FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY package*.json ./
RUN npm install

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
HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost:$PORT/ || exit 1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := types.Facts{
				HasLockfile: false,
			}

			gen := DockerfileGenerator{}
			dockerfile, err := gen.Dockerfile(&facts)
			if err != nil {
				t.Fatalf("Dockerfile() error = %v", err)
			}

			if dockerfile != tt.expected {
				t.Errorf("Dockerfile() = %v, want %v", dockerfile, tt.expected)
			}
		})
	}
}

func TestDockerfileGeneration(t *testing.T) {
	tests := []struct {
		name        string
		hasLockfile bool
		wantCmd     string
	}{
		{
			name:        "with package-lock.json",
			hasLockfile: true,
			wantCmd:     "RUN npm ci --silent",
		},
		{
			name:        "without package-lock.json",
			hasLockfile: false,
			wantCmd:     "RUN npm install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory
			dir := t.TempDir()

			// Create package.json
			if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
				t.Fatal(err)
			}

			// Create package-lock.json if needed
			if tt.hasLockfile {
				if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(`{"name":"test"}`), 0644); err != nil {
					t.Fatal(err)
				}
			}

			// Create facts with HasLockfile field
			facts := &types.Facts{
				Ports:       []int{80},
				HasLockfile: tt.hasLockfile,
			}

			// Generate Dockerfile
			dockerfile, err := DockerfileGenerator{}.Dockerfile(facts)
			assert.NoError(t, err)
			assert.Contains(t, dockerfile, tt.wantCmd)
		})
	}
}
