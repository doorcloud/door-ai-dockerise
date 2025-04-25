package detect

import (
	"testing"
	"testing/fstest"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantRule Rule
		wantErr  bool
	}{
		{
			name: "spring boot maven project",
			files: map[string]string{
				"pom.xml": "<project></project>",
			},
			wantRule: Rule{
				Name: "spring-boot",
				Tool: "maven",
			},
			wantErr: false,
		},
		{
			name:     "empty project",
			files:    map[string]string{},
			wantRule: Rule{},
			wantErr:  false,
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
			gotRule, err := Detect(fsys)
			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRule != tt.wantRule {
				t.Errorf("Detect() = %v, want %v", gotRule, tt.wantRule)
			}
		})
	}
}
