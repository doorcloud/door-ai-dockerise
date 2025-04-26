package e2e

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReactSpec(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "react-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a minimal React application
	createMinimalReactApp(t, tempDir)

	// TODO: Initialize pipeline with actual implementations
	// p := pipeline.NewPipeline(...)

	// For now, we'll create a simple Dockerfile directly
	dockerfile := `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`

	// Write the Dockerfile
	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	err = os.WriteFile(dockerfilePath, []byte(dockerfile), 0644)
	require.NoError(t, err)

	// Build the Docker image
	imageName := "react-test-image"
	err = buildDockerImage(t, tempDir, imageName)
	require.NoError(t, err)
	defer removeDockerImage(t, imageName)

	// Run the container
	containerID, err := runDockerContainer(t, imageName)
	require.NoError(t, err)
	defer stopDockerContainer(t, containerID)

	// Wait for the container to be ready
	time.Sleep(2 * time.Second)

	// Test the root endpoint
	resp, err := http.Get("http://localhost:3000")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func createMinimalReactApp(t *testing.T, dir string) {
	// Create package.json
	packageJSON := `{
        "name": "react-test",
        "version": "0.1.0",
        "private": true,
        "dependencies": {
            "react": "^18.2.0",
            "react-dom": "^18.2.0",
            "react-scripts": "5.0.1"
        },
        "scripts": {
            "start": "react-scripts start",
            "build": "react-scripts build"
        }
    }`
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0644)
	require.NoError(t, err)

	// Create src/index.js
	srcDir := filepath.Join(dir, "src")
	err = os.MkdirAll(srcDir, 0755)
	require.NoError(t, err)

	indexJS := `import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
    <React.StrictMode>
        <h1>Hello, World!</h1>
    </React.StrictMode>
);`
	err = os.WriteFile(filepath.Join(srcDir, "index.js"), []byte(indexJS), 0644)
	require.NoError(t, err)

	// Create src/index.css
	indexCSS := `body {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
			'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
			sans-serif;
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
	}

	code {
		font-family: source-code-pro, Menlo, Monaco, Consolas, 'Courier New',
			monospace;
	}`
	err = os.WriteFile(filepath.Join(srcDir, "index.css"), []byte(indexCSS), 0644)
	require.NoError(t, err)

	// Create public/index.html
	publicDir := filepath.Join(dir, "public")
	err = os.MkdirAll(publicDir, 0755)
	require.NoError(t, err)

	indexHTML := `<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8" />
        <title>React Test</title>
    </head>
    <body>
        <div id="root"></div>
    </body>
</html>`
	err = os.WriteFile(filepath.Join(publicDir, "index.html"), []byte(indexHTML), 0644)
	require.NoError(t, err)
}

func buildDockerImage(t *testing.T, contextDir string, imageName string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = contextDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDockerContainer(t *testing.T, imageName string) (string, error) {
	cmd := exec.Command("docker", "run", "-d", "-p", "3000:3000", imageName)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func stopDockerContainer(t *testing.T, containerID string) error {
	cmd := exec.Command("docker", "stop", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func removeDockerImage(t *testing.T, imageName string) error {
	cmd := exec.Command("docker", "rmi", "-f", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
