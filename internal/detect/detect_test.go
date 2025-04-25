package detect

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

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
			fsys := fstest.MapFS{}
			for _, f := range tt.files {
				fsys[f] = &fstest.MapFile{}
			}

			got, err := Detect(fsys)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
