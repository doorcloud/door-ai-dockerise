package stack

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	ctx := context.Background()
	fsys := os.DirFS("testdata")

	rule, err := Detect(ctx, fsys)
	assert.NoError(t, err)
	assert.NotNil(t, rule)

	// TODO: Add more specific tests
}
