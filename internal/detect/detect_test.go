package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/stretchr/testify/assert"
)

type testFile struct {
	path    string
	content string
}

func writeFiles(dir string, files []testFile) error {
	for _, f := range files {
		path := filepath.Join(dir, f.path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		content := f.content
		if content == "" {
			content = "{}" // Default to empty JSON for empty files
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    []testFile
		expected types.RuleInfo
		wantErr  bool
	}{
		{
			name: "spring boot with maven",
			files: []testFile{
				{path: "pom.xml"},
			},
			expected: types.RuleInfo{
				Name: "spring-boot",
				Tool: "maven",
			},
		},
		{
			name: "spring boot with gradle",
			files: []testFile{
				{path: "gradlew"},
			},
			expected: types.RuleInfo{
				Name: "spring-boot",
				Tool: "gradle",
			},
		},
		{
			name: "node with pnpm",
			files: []testFile{
				{path: "package.json"},
				{path: "pnpm-lock.yaml"},
			},
			expected: types.RuleInfo{
				Name: "node",
				Tool: "pnpm",
			},
		},
		{
			name: "react with npm",
			files: []testFile{
				{path: "package.json", content: `{"dependencies": {"react": "18.2.0", "react-scripts": "5.0.1"}}`},
				{path: "src/index.js"},
			},
			expected: types.RuleInfo{
				Name: "react",
				Tool: "npm",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := writeFiles(dir, tt.files); err != nil {
				t.Fatalf("writeFiles() error = %v", err)
			}

			got, err := Detect(os.DirFS(dir))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
