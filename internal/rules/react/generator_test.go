package react

import (
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	g := NewGenerator()

	facts := core.Facts{
		StackType: "react",
		BuildTool: "npm",
		Port:      3000,
	}

	dockerfile, err := g.Generate(facts)
	require.NoError(t, err)
	require.Contains(t, dockerfile, "FROM node:14")
	require.Contains(t, dockerfile, "EXPOSE 3000")
	require.Contains(t, dockerfile, "npm run build")
}
