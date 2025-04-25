package rules

import (
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
)

func TestDetectStack(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantRule *detect.Rule
		wantErr  error
	}{
		{
			name: "spring boot maven",
			files: map[string]string{
				"pom.xml": "<project></project>",
			},
			wantRule: &detect.Rule{
				Name: "spring-boot",
				Tool: "maven",
			},
			wantErr: nil,
		},
		{
			name: "spring boot gradle",
			files: map[string]string{
				"gradlew": "#!/bin/sh",
			},
			wantRule: &detect.Rule{
				Name: "spring-boot",
				Tool: "gradle",
			},
			wantErr: nil,
		},
		{
			name:     "no match",
			files:    map[string]string{},
			wantRule: nil,
			wantErr:  ErrUnknownStack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test filesystem
			fsys := fstest.MapFS{}
			for path, content := range tt.files {
				fsys[path] = &fstest.MapFile{Data: []byte(content)}
			}

			// Run detection
			gotRule, gotErr := DetectStack(fsys)
			if gotErr != tt.wantErr {
				t.Errorf("DetectStack() error = %v, want %v", gotErr, tt.wantErr)
			}
			if tt.wantRule != nil && *gotRule != *tt.wantRule {
				t.Errorf("DetectStack() rule = %v, want %v", gotRule, tt.wantRule)
			}
		})
	}
}
