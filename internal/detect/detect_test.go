package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeFiles(dir string, files []string) error {
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			return err
		}
	}
	return nil
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected RuleInfo
		wantErr  bool
	}{
		{
			name:  "spring boot with maven",
			files: []string{"pom.xml"},
			expected: RuleInfo{
				Name: "spring-boot",
				Tool: "maven",
			},
		},
		{
			name:  "spring boot with gradle",
			files: []string{"gradlew"},
			expected: RuleInfo{
				Name: "spring-boot",
				Tool: "gradle",
			},
		},
		{
			name:  "node with pnpm",
			files: []string{"pnpm-lock.yaml"},
			expected: RuleInfo{
				Name: "node",
				Tool: "pnpm",
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
