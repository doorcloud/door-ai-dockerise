package react

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactsDetector(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	fmt.Printf("Current working directory: %s\n", wd)
	root := filepath.Join(wd, "..", "..", "..", "testdata/react-min")
	fmt.Printf("Test fixture directory: %s\n", root)
	fd := FactsDetector{}
	require.True(t, fd.Detect(os.DirFS(root)))
	facts := fd.Facts(os.DirFS(root))
	assert.Equal(t, "React", facts["framework"])
	assert.Contains(t, facts["build_cmd"], "build")
}
