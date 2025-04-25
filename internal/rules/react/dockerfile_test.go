package react

import (
	"testing"

	"github.com/doorcloud/door-ai-dockerise/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerfileGenerator(t *testing.T) {
	gen := DockerfileGenerator{}
	facts := &types.Facts{Framework: "React"} // minimal
	df, err := gen.Dockerfile(facts)
	require.NoError(t, err)
	assert.Contains(t, df, "node:18-alpine")
	assert.Contains(t, df, "COPY --from=build")
}
