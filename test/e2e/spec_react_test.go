package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers"
)

func TestReactSpec(t *testing.T) {
	// Create a temporary directory for the test
	dir, err := os.MkdirTemp("", "react-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create package.json
	pkgJSON := `{
		"name": "react-test",
		"version": "0.1.0",
		"private": true,
		"dependencies": {
			"@testing-library/jest-dom": "^5.17.0",
			"@testing-library/react": "^13.4.0",
			"@testing-library/user-event": "^13.5.0",
			"react": "^18.2.0",
			"react-dom": "^18.2.0",
			"react-scripts": "5.0.1",
			"web-vitals": "^2.1.4"
		},
		"scripts": {
			"start": "react-scripts start",
			"build": "react-scripts build",
			"test": "react-scripts test",
			"eject": "react-scripts eject"
		},
		"eslintConfig": {
			"extends": [
				"react-app",
				"react-app/jest"
			]
		},
		"browserslist": {
			"production": [
				">0.2%",
				"not dead",
				"not op_mini all"
			],
			"development": [
				"last 1 chrome version",
				"last 1 firefox version",
				"last 1 safari version"
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create src directory
	if err := os.Mkdir(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	// Create index.js
	indexJS := `import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <h1>Hello, World!</h1>
  </React.StrictMode>
);`
	if err := os.WriteFile(filepath.Join(dir, "src", "index.js"), []byte(indexJS), 0o644); err != nil {
		t.Fatalf("Failed to write index.js: %v", err)
	}

	// Create index.css
	indexCSS := `body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}`
	if err := os.WriteFile(filepath.Join(dir, "src", "index.css"), []byte(indexCSS), 0o644); err != nil {
		t.Fatalf("Failed to write index.css: %v", err)
	}

	// Create public directory
	if err := os.Mkdir(filepath.Join(dir, "public"), 0o755); err != nil {
		t.Fatalf("Failed to create public directory: %v", err)
	}

	// Create index.html
	indexHTML := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>React App</title>
  </head>
  <body>
    <div id="root"></div>
  </body>
</html>`
	if err := os.WriteFile(filepath.Join(dir, "public", "index.html"), []byte(indexHTML), 0o644); err != nil {
		t.Fatalf("Failed to write index.html: %v", err)
	}

	// Create Dockerfile
	dockerfile := `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN apk add --no-cache python3 make g++ \
    && npm install \
    && apk del python3 make g++
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0o644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	// Create Docker verifier with logging
	docker, err := verifiers.NewDocker(verifiers.Options{
		Socket:  "unix:///var/run/docker.sock",
		LogSink: os.Stdout,
		Timeout: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Failed to create Docker verifier: %v", err)
	}

	// Verify the Dockerfile
	ctx := context.Background()
	if err := docker.Verify(ctx, dir, "Dockerfile"); err != nil {
		t.Fatalf("Failed to verify Dockerfile: %v", err)
	}
}
