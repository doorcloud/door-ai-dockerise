package detectors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestReactDetector(t *testing.T) {
	// Create test directories
	tmpDir, err := os.MkdirTemp("", "react-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a React project directory
	reactDir := filepath.Join(tmpDir, "react-project")
	if err := os.MkdirAll(reactDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create package.json with React and build tool dependencies
	packageJSON := `{
		"name": "test-react-app",
		"dependencies": {
			"react": "^18.0.0",
			"react-scripts": "5.0.1"
		}
	}`
	if err := os.WriteFile(filepath.Join(reactDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a non-React project directory
	nonReactDir := filepath.Join(tmpDir, "non-react-project")
	if err := os.MkdirAll(nonReactDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create package.json without React dependency
	nonReactPackageJSON := `{
		"name": "test-app",
		"dependencies": {
			"express": "^4.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(nonReactDir, "package.json"), []byte(nonReactPackageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		dir      string
		wantInfo core.StackInfo
		wantErr  bool
	}{
		{
			name: "react project",
			dir:  reactDir,
			wantInfo: core.StackInfo{
				Name: "react",
				Meta: map[string]string{
					"framework": "react",
				},
			},
			wantErr: false,
		},
		{
			name:     "non-react project",
			dir:      nonReactDir,
			wantInfo: core.StackInfo{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReact()
			got, err := r.Detect(context.Background(), tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("React.Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.wantInfo.Name {
				t.Errorf("React.Detect() = %v, want %v", got, tt.wantInfo)
			}
		})
	}
}
