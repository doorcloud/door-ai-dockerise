package e2e

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/react"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/adapters/rules/springboot"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	dockermock "github.com/doorcloud/door-ai-dockerise/drivers/docker/mock"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline"
	llmmock "github.com/doorcloud/door-ai-dockerise/providers/llm/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReactProject(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping E2E test. Set DG_E2E=1 to run.")
	}

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create a mock React project
	packageJson := `{
		"name": "test-react-app",
		"version": "1.0.0",
		"scripts": {
			"start": "react-scripts start",
			"build": "react-scripts build"
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJson), 0o644))

	// Run the generator
	output, err := runGenerator(tempDir)
	require.NoError(t, err)

	// Verify the generated Dockerfile
	expected := `FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./
RUN npm install --production

EXPOSE 3001
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/ || exit 1

CMD ["npm", "start"]`

	require.Equal(t, expected, output)
}

func runGenerator(projectDir string) (string, error) {
	// Create mock LLM
	mockLLM := llmmock.NewMockClient()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline with mock components
	p := v2.NewPipeline(
		v2.WithDetectors(
			react.NewReactDetector(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(dockermock.NewMockDriver()),
		v2.WithMaxRetries(3),
		v2.WithLogSink(&logBuf),
	)

	// Create test context
	ctx := context.Background()

	// Run the pipeline
	if err := p.Run(ctx, projectDir); err != nil {
		return "", err
	}

	// Read the generated Dockerfile
	dockerfilePath := filepath.Join(projectDir, "Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func TestReactIntegration(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping integration test; set DG_E2E=1 to run")
	}

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline with mock components
	p := v2.NewPipeline(
		v2.WithDetectors(
			react.NewReactDetector(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(dockermock.NewMockDriver()),
		v2.WithMaxRetries(3),
		v2.WithLogSink(&logBuf),
	)

	// Create test context
	ctx := context.Background()

	// Get absolute path to test project
	projectPath, err := filepath.Abs("testdata/react-project")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the pipeline
	if err := p.Run(ctx, projectPath); err != nil {
		t.Errorf("Pipeline.Run() error = %v", err)
	}

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Errorf("Dockerfile was not created at %s", dockerfilePath)
	}

	// Verify log output
	logOutput := logBuf.String()
	assert.True(t, strings.Contains(logOutput, "detector=react found=true"), "Expected React detector log line")
	assert.False(t, strings.Contains(logOutput, "detector=springboot found=true"), "Unexpected SpringBoot detector log line")
}

func TestIntegration_React(t *testing.T) {
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
	}`), 0o644)
	assert.NoError(t, err)

	// Create src/index.js
	err = os.MkdirAll(filepath.Join(dir, "src"), 0o755)
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
	`), 0o644)
	assert.NoError(t, err)

	// Create src/index.css
	err = os.WriteFile(filepath.Join(dir, "src", "index.css"), []byte(`
		body {
			margin: 0;
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
			-webkit-font-smoothing: antialiased;
		}
	`), 0o644)
	assert.NoError(t, err)

	// Create public/index.html
	err = os.MkdirAll(filepath.Join(dir, "public"), 0o755)
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
	`), 0o644)
	assert.NoError(t, err)

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline
	p := v2.NewPipeline(
		v2.WithDetectors(
			react.NewReactDetector(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(dockermock.NewMockDriver()),
		v2.WithMaxRetries(3),
		v2.WithLogSink(&logBuf),
	)

	// Run the pipeline
	ctx := context.Background()
	err = p.Run(ctx, dir)
	assert.NoError(t, err)

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Errorf("Dockerfile was not created at %s", dockerfilePath)
	}

	// Verify log output
	logOutput := logBuf.String()
	assert.True(t, strings.Contains(logOutput, "detector=react found=true"), "Expected React detector log line")
	assert.False(t, strings.Contains(logOutput, "detector=springboot found=true"), "Unexpected SpringBoot detector log line")
}

func TestReactSpecV2(t *testing.T) {
	if os.Getenv("DG_E2E") == "" {
		t.Skip("Skipping integration test. Set DG_E2E=1 to run.")
	}

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline with mock components
	p := v2.NewPipeline(
		v2.WithDetectors(
			react.NewReactDetector(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(dockermock.NewMockDriver()),
		v2.WithMaxRetries(3),
		v2.WithLogSink(&logBuf),
	)

	// Create test context
	ctx := context.Background()

	// Get absolute path to test project
	projectPath, err := filepath.Abs("testdata/react")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the pipeline
	if err := p.Run(ctx, projectPath); err != nil {
		t.Errorf("Pipeline.Run() error = %v", err)
	}

	// Verify Dockerfile was created
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Errorf("Dockerfile was not created at %s", dockerfilePath)
	}

	// Verify log output
	logOutput := logBuf.String()
	assert.True(t, strings.Contains(logOutput, "detector=react found=true"), "Expected React detector log line")
	assert.False(t, strings.Contains(logOutput, "detector=springboot found=true"), "Unexpected SpringBoot detector log line")
}
