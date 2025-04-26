package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/log/mock"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	"github.com/stretchr/testify/assert"
)

func TestReactSpec(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test; set DG_E2E=1 to run")
	}

	// Create a temporary directory for the test
	dir := t.TempDir()

	// Create package.json
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"name": "test-app",
		"version": "1.0.0",
		"scripts": {
			"start": "react-scripts start",
			"build": "react-scripts build"
		},
		"dependencies": {
			"react": "^18.2.0",
			"react-dom": "^18.2.0",
			"react-scripts": "5.0.1"
		}
	}`), 0644)
	assert.NoError(t, err)

	// Create src/index.js
	err = os.MkdirAll(filepath.Join(dir, "src"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "src", "index.js"), []byte(`
		import React from 'react';
		import ReactDOM from 'react-dom/client';
		import './index.css';

		const root = ReactDOM.createRoot(document.getElementById('root'));
		root.render(
			<React.StrictMode>
				<h1>Hello, World!</h1>
			</React.StrictMode>
		);
	`), 0644)
	assert.NoError(t, err)

	// Create src/index.css
	err = os.WriteFile(filepath.Join(dir, "src", "index.css"), []byte(`
		body {
			margin: 0;
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
			-webkit-font-smoothing: antialiased;
		}
	`), 0644)
	assert.NoError(t, err)

	// Create public/index.html
	err = os.MkdirAll(filepath.Join(dir, "public"), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "public", "index.html"), []byte(`
		<!DOCTYPE html>
		<html>
			<head>
				<meta charset="utf-8" />
				<title>Test App</title>
			</head>
			<body>
				<div id="root"></div>
			</body>
		</html>
	`), 0644)
	assert.NoError(t, err)

	// Create Dockerfile
	dockerfile := `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`
	err = os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644)
	assert.NoError(t, err)

	// Create Docker engine
	engine, err := docker.NewEngine()
	assert.NoError(t, err)

	// Create build input
	input := core.BuildInput{
		ContextTar: createContextTar(t, dir),
		Dockerfile: dockerfile,
	}

	// Build the image
	log := mock.New()
	_, err = engine.Build(context.Background(), input, log)
	assert.NoError(t, err)

	// Verify logs
	entries := log.Entries()
	assert.NotEmpty(t, entries, "Expected log entries, got none")
}

func createContextTar(t *testing.T, dir string) io.Reader {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	// Walk the directory and add files to the tar
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Set the path relative to the root
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
	assert.NoError(t, err)

	return &buf
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
	assert.NoError(t, err)

	// Create src/index.js
	srcDir := filepath.Join(dir, "src")
	err = os.MkdirAll(srcDir, 0755)
	assert.NoError(t, err)

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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	// Create public/index.html
	publicDir := filepath.Join(dir, "public")
	err = os.MkdirAll(publicDir, 0755)
	assert.NoError(t, err)

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
	assert.NoError(t, err)
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
