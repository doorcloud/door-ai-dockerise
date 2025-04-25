package springboot

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
		name  string
		files []string
		want  bool
	}{
		{
			name:  "maven project",
			files: []string{"pom.xml"},
			want:  true,
		},
		{
			name:  "gradle project",
			files: []string{"gradlew"},
			want:  true,
		},
		{
			name:  "gradle kotlin project",
			files: []string{"build.gradle.kts"},
			want:  true,
		},
		{
			name:  "not a spring boot project",
			files: []string{"package.json"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := writeFiles(dir, tt.files); err != nil {
				t.Fatalf("writeFiles() error = %v", err)
			}

			got := Rule{}.Detect(os.DirFS(dir))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFacts(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  map[string]any
	}{
		{
			name:  "maven project",
			files: []string{"pom.xml"},
			want: map[string]any{
				"language":   "Java",
				"framework":  "Spring Boot",
				"build_tool": "maven",
				"build_cmd":  "mvn clean package",
				"start_cmd":  "java -jar target/*.jar",
				"artifact":   "target/*.jar",
				"ports":      []int{8080},
			},
		},
		{
			name:  "gradle project",
			files: []string{"gradlew"},
			want: map[string]any{
				"language":   "Java",
				"framework":  "Spring Boot",
				"build_tool": "gradle",
				"build_cmd":  "mvn clean package",
				"start_cmd":  "java -jar target/*.jar",
				"artifact":   "target/*.jar",
				"ports":      []int{8080},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := writeFiles(dir, tt.files); err != nil {
				t.Fatalf("writeFiles() error = %v", err)
			}

			got := Rule{}.Facts(os.DirFS(dir))
			assert.Equal(t, tt.want, got)
		})
	}
}
