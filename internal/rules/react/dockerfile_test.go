package react

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerfileGenerator(t *testing.T) {
	gen := DockerfileGenerator{}
	facts := &types.Facts{
		Framework: "React",
		Ports:     []int{3000},
	}
	df, err := gen.Dockerfile(facts)
	require.NoError(t, err)
	assert.Contains(t, df, "node:18-alpine")
	assert.Contains(t, df, "COPY --from=build")
	assert.Contains(t, df, "EXPOSE 3000")
}

func TestDockerfileGeneration(t *testing.T) {
	tests := []struct {
		name        string
		hasLockfile bool
		wantCmd     string
	}{
		{
			name:        "with package-lock.json",
			hasLockfile: true,
			wantCmd:     "RUN npm ci --silent",
		},
		{
			name:        "without package-lock.json",
			hasLockfile: false,
			wantCmd:     "RUN npm install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory
			dir := t.TempDir()

			// Create package.json
			if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
				t.Fatal(err)
			}

			// Create package-lock.json if needed
			if tt.hasLockfile {
				if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(`{"name":"test"}`), 0644); err != nil {
					t.Fatal(err)
				}
			}

			// Create facts with HasLockfile field
			facts := &types.Facts{
				Ports:       []int{80},
				HasLockfile: tt.hasLockfile,
			}

			// Generate Dockerfile
			dockerfile, err := DockerfileGenerator{}.Dockerfile(facts)
			assert.NoError(t, err)
			assert.Contains(t, dockerfile, tt.wantCmd)
		})
	}
}
