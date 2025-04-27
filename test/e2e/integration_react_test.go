package e2e

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors"
	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/springboot"
	"github.com/doorcloud/door-ai-dockerise/adapters/facts"
	"github.com/doorcloud/door-ai-dockerise/adapters/generate"
	"github.com/doorcloud/door-ai-dockerise/core/mock"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker"
	v2 "github.com/doorcloud/door-ai-dockerise/pipeline/v2"
	"github.com/stretchr/testify/assert"
)

func TestReactProject(t *testing.T) {
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
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(docker.NewMockDriver()),
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
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(docker.NewMockDriver()),
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

	// Create mock LLM
	mockLLM := mock.NewMockLLM()

	// Create buffer for log output
	var logBuf bytes.Buffer

	// Create pipeline
	p := v2.NewPipeline(
		v2.WithDetectors(
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(docker.NewMockDriver()),
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
			detectors.NewReact(),
			springboot.NewSpringBootDetector(),
		),
		v2.WithFactProviders(
			facts.NewStatic(),
		),
		v2.WithGenerator(generate.NewLLM(mockLLM)),
		v2.WithDockerDriver(docker.NewMockDriver()),
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
