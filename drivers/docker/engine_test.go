package docker

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/doorcloud/door-ai-dockerise/core"
	"github.com/stretchr/testify/assert"
)

type fakeDriver struct {
	fail bool
}

func (f *fakeDriver) Build(ctx context.Context, in core.BuildInput, w io.Writer) (core.ImageRef, error) {
	if f.fail {
		return core.ImageRef{}, assert.AnError
	}
	_, err := w.Write([]byte("built\n"))
	if err != nil {
		return core.ImageRef{}, err
	}
	return core.ImageRef{Name: "test-image:latest"}, nil
}

func TestEngine_Build(t *testing.T) {
	tests := []struct {
		name    string
		driver  *fakeDriver
		wantErr bool
	}{
		{
			name:    "successful build",
			driver:  &fakeDriver{fail: false},
			wantErr: false,
		},
		{
			name:    "failed build",
			driver:  &fakeDriver{fail: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create a minimal build input
			in := core.BuildInput{
				ContextTar: bytes.NewReader([]byte("test")),
				Dockerfile: "FROM alpine\n",
			}

			// Run the build
			ref, err := tt.driver.Build(ctx, in, &buf)

			// Check the results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, ref.Name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "test-image:latest", ref.Name)
				assert.Contains(t, buf.String(), "built")
			}
		})
	}
}
