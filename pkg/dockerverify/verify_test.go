package dockerverify

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestVerify(t *testing.T) {
	// Create a minimal test context
	fsys := fstest.MapFS{
		"go.mod": &fstest.MapFile{
			Data: []byte(`module test
go 1.21
`),
		},
	}

	// Test Dockerfile verification
	dockerfile := `FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o app
CMD ["./app"]`

	err := Verify(context.Background(), fsys, dockerfile)
	if err != nil {
		t.Errorf("Verify() error = %v", err)
	}
}
