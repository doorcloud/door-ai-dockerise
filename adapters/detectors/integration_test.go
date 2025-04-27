package detectors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	testCases := []struct {
		name     string
		testPath string
	}{
		{"spring_kotlin", "spring/kotlin_gradle_demo"},
		{"spring_maven", "spring/petclinic_maven"},
		{"react", "react/cra_template"},
		{"node", "node/express_hello"},
		{"plain_java", "negative/plain_java"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := filepath.Join("../../testdata/e2e", tc.testPath)

			// Read expected stack from EXPECTED_STACK file
			expectedBytes, err := os.ReadFile(filepath.Join(testDir, "EXPECTED_STACK"))
			require.NoError(t, err)
			expected := strings.TrimSpace(string(expectedBytes))

			// Create detector
			detector := NewFanoutDetector()

			// Run detection
			fsys := os.DirFS(testDir)
			info, found, err := detector.Detect(context.Background(), fsys, nil)
			require.NoError(t, err)

			if expected == "unknown" {
				assert.False(t, found, "Expected no stack to be detected")
				return
			}

			assert.True(t, found, "Expected stack to be detected")
			assert.Equal(t, expected, info.Name, "Stack name mismatch")
		})
	}
}
