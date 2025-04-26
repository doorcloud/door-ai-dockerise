package react

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
)

func TestReactDetector(t *testing.T) {
	// Get the absolute path to the fixtures directory
	fixturesDir := filepath.Join("test", "e2e", "fixtures", "react-min")
	absPath, err := filepath.Abs(fixturesDir)
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
				Name: "react",
				Meta: map[string]string{
					"framework": "react",
					"buildTool": "npm",
				},
			},
			wantErr: false,
		},
		{
			name:     "non-existent path",
			path:     "/nonexistent/path",
			wantInfo: core.StackInfo{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewReactDetector()
			got, err := d.Detect(context.Background(), tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReactDetector.Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.wantInfo.Name {
				t.Errorf("ReactDetector.Detect() = %v, want %v", got, tt.wantInfo)
			}
			if got.Name != "" {
				if got.Meta["framework"] != tt.wantInfo.Meta["framework"] {
					t.Errorf("ReactDetector.Detect() framework = %v, want %v", got.Meta["framework"], tt.wantInfo.Meta["framework"])
				}
				if got.Meta["buildTool"] != tt.wantInfo.Meta["buildTool"] {
					t.Errorf("ReactDetector.Detect() buildTool = %v, want %v", got.Meta["buildTool"], tt.wantInfo.Meta["buildTool"])
				}
			}
		})
	}
}
