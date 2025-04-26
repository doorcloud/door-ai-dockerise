package react

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReactDetector(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
	}{
		{
			name: "root package.json positive",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				pkg := `{
					"dependencies": {
						"react": "18.2.0",
						"react-scripts": "5.0.1"
					}
				}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))
				return dir
			},
			expected: true,
		},
		{
			name: "nested package.json positive",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "examples", "react"), 0755))
				pkg := `{
					"dependencies": {
						"react": "18.2.0",
						"vite": "4.0.0"
					}
				}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "examples", "react", "package.json"), []byte(pkg), 0644))
				return dir
			},
			expected: true,
		},
		{
			name: "missing build tool negative",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				pkg := `{
					"dependencies": {
						"react": "18.2.0"
					}
				}`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644))
				return dir
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			detected := (&ReactDetector{}).Detect(os.DirFS(dir))
			assert.Equal(t, tt.expected, detected)
		})
	}
}
