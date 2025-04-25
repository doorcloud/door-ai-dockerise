package rules

import (
	"testing"
	"testing/fstest"

	"github.com/doorcloud/door-ai-dockerise/internal/detect"
	"github.com/doorcloud/door-ai-dockerise/internal/detect/springboot"
)

func TestRegistry_Detect(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantRule detect.Rule
		wantOk   bool
	}{
		{
			name: "spring boot maven",
			files: map[string]string{
				"pom.xml": "<project></project>",
			},
			wantRule: detect.Rule{
				Name: "spring-boot",
				Tool: "maven",
			},
			wantOk: true,
		},
		{
			name: "spring boot gradle",
			files: map[string]string{
				"gradlew": "#!/bin/sh",
			},
			wantRule: detect.Rule{
				Name: "spring-boot",
				Tool: "gradle",
			},
			wantOk: true,
		},
		{
			name:     "no match",
			files:    map[string]string{},
			wantRule: detect.Rule{},
			wantOk:   false,
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
			reg := NewRegistry()
			reg.Register(&springboot.Detector{})
			gotRule, gotOk := reg.Detect(fsys)
			if gotOk != tt.wantOk {
				t.Errorf("Registry.Detect() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotRule != tt.wantRule {
				t.Errorf("Registry.Detect() rule = %v, want %v", gotRule, tt.wantRule)
			}
		})
	}
}
