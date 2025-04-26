package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "spec-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name        string
		content     string
		expectError bool
		expected    *core.Spec
	}{
		{
			name: "valid spec with all fields",
			content: `
language: javascript
framework: node
version: "18"
buildTool: npm
params:
  port: "3000"
  env: "production"
`,
			expectError: false,
			expected: &core.Spec{
				Language:  "javascript",
				Framework: "node",
				Version:   "18",
				BuildTool: "npm",
				Params:    map[string]string{"port": "3000", "env": "production"},
			},
		},
		{
			name: "valid spec with minimal fields",
			content: `
language: python
framework: django
`,
			expectError: false,
			expected: &core.Spec{
				Language:  "python",
				Framework: "django",
				Version:   "",
				BuildTool: "",
				Params:    map[string]string{},
			},
		},
		{
			name: "invalid yaml syntax",
			content: `
language: javascript
framework: node
version: "18"
buildTool: npm
invalid: yaml: here
`,
			expectError: true,
			expected:    nil,
		},
		{
			name: "missing required fields",
			content: `
version: "18"
buildTool: npm
`,
			expectError: true,
			expected:    nil,
		},
		{
			name: "empty spec",
			content: `
language: ""
framework: ""
version: ""
buildTool: ""
params: {}
`,
			expectError: true,
			expected:    nil,
		},
		{
			name: "invalid params type",
			content: `
language: javascript
framework: node
params: "invalid"
`,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the test content
			filePath := filepath.Join(tempDir, tt.name+".yaml")
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load the spec
			spec, err := Load(filePath)

			// Check error expectations
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, spec)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, spec)
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Try to load a non-existent file
	spec, err := Load("nonexistent.yaml")
	assert.Error(t, err)
	assert.Nil(t, spec)
}

func TestLoad_InvalidExtension(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "spec-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a file with invalid extension
	filePath := filepath.Join(tempDir, "spec.txt")
	err = os.WriteFile(filePath, []byte("language: javascript"), 0644)
	require.NoError(t, err)

	// Try to load the file
	spec, err := Load(filePath)
	assert.Error(t, err)
	assert.Nil(t, spec)
}
