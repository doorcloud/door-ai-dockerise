package react

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestReactDetector(t *testing.T) {
	// Get the absolute path to the testdata directory
	testdataDir := filepath.Join("..", "..", "..", "testdata", "react-min")
	absPath, err := filepath.Abs(testdataDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantInfo core.StackInfo
		wantErr  bool
	}{
		{
			name: "react project",
			path: absPath,
			wantInfo: core.StackInfo{
				Name:          "react",
				BuildTool:     "npm",
				DetectedFiles: []string{"package.json"},
			},
			wantErr: false,
		},
		{
			name:     "non-existent path",
			path:     "/nonexistent/path",
			wantInfo: core.StackInfo{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewReactDetector()
			fsys := os.DirFS(tt.path)
			got, found, err := d.Detect(context.Background(), fsys, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReactDetector.Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.wantInfo.Name {
				t.Errorf("ReactDetector.Detect() = %v, want %v", got.Name, tt.wantInfo.Name)
			}
			if got.BuildTool != tt.wantInfo.BuildTool {
				t.Errorf("ReactDetector.Detect() = %v, want %v", got.BuildTool, tt.wantInfo.BuildTool)
			}
			if found != (tt.wantInfo.Name != "") {
				t.Errorf("ReactDetector.Detect() found = %v, want %v", found, tt.wantInfo.Name != "")
			}
		})
	}
}
