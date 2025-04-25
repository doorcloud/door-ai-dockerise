package springboot

import (
	"testing"
	"testing/fstest"
)

func TestDetector_Detect(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string
		wantHasWrapper bool
	}{
		{
			name: "has maven wrapper",
			files: map[string]string{
				"mvnw":               "#!/bin/sh",
				"mvnw.cmd":           "@REM",
				".mvn/wrapper/a.jar": "content",
			},
			wantHasWrapper: true,
		},
		{
			name: "no maven wrapper",
			files: map[string]string{
				"pom.xml": "<project>",
			},
			wantHasWrapper: false,
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
			d := &Detector{}
			if err := d.Detect(fsys); err != nil {
				t.Errorf("Detect() error = %v", err)
				return
			}

			if d.HasMavenWrapper != tt.wantHasWrapper {
				t.Errorf("HasMavenWrapper = %v, want %v", d.HasMavenWrapper, tt.wantHasWrapper)
			}
		})
	}
}
