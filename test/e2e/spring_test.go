package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/adapters/detectors/spring"
	"github.com/doorcloud/door-ai-dockerise/adapters/generator"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers/docker"
	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_SpringMatrix(t *testing.T) {
	if os.Getenv("DG_E2E") != "1" {
		t.Skip("Skipping E2E test. Set DG_E2E=1 to run.")
	}

	// Get project root directory
	wd, err := os.Getwd()
	require.NoError(t, err)
	rootDir := filepath.Dir(filepath.Dir(wd))

	tests := []struct {
		name     string
		project  string
		port     int
		expected string
	}{
		{
			name:     "Petclinic Maven project",
			project:  "testdata/e2e/spring/petclinic",
			port:     8080,
			expected: "FROM maven:3.8-openjdk-17 AS build",
		},
		{
			name:     "Gradle Kotlin project",
			project:  "testdata/e2e/spring/gradle-kts",
			port:     8080,
			expected: "FROM gradle:7.5-jdk17 AS build",
		},
		{
			name:     "Nested Maven project",
			project:  "testdata/e2e/spring/nested",
			port:     8080,
			expected: "FROM maven:3.8-openjdk-17 AS build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "spring-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Copy test project
			err = copyDir(filepath.Join(rootDir, tt.project), tempDir)
			require.NoError(t, err)

			// Create detector
			detector := spring.NewSpringBootDetectorV3()

			// Create generator
			gen := generator.NewTemplateGenerator()

			// Create verifier
			verifier := docker.NewVerifier()

			// Create log sink
			logSink := &core.StringLogSink{}

			// Detect stack
			fsys := os.DirFS(tempDir)
			info, found, err := detector.Detect(context.Background(), fsys, logSink)
			require.NoError(t, err)
			require.True(t, found)

			// Generate Dockerfile
			facts := core.Facts{
				StackType: info.Name,
				BuildTool: info.BuildTool,
				Port:      info.Port,
			}
			dockerfile, err := gen.Generate(context.Background(), facts)
			require.NoError(t, err)
			assert.Contains(t, dockerfile, tt.expected)

			// Write Dockerfile
			err = os.WriteFile(filepath.Join(tempDir, "Dockerfile"), []byte(dockerfile), 0o644)
			require.NoError(t, err)

			// Verify Dockerfile
			err = verifier.Verify(context.Background(), tempDir, "Dockerfile", tt.port, os.Stdout)
			require.NoError(t, err)
		})
	}
}
