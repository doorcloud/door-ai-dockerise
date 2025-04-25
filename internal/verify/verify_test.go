package verify

import (
	"context"
	"os/exec"
	"testing"
	"testing/fstest"
)

func TestVerify(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}

	// Create a test filesystem
	fsys := fstest.MapFS{
		"main.go": &fstest.MapFile{
			Data: []byte(`package main

func main() {
	println("Hello, World!")
}`),
		},
	}

	// Test a valid Dockerfile
	dockerfile := `FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o app
CMD ["./app"]`

	err := Verify(context.Background(), fsys, dockerfile)
	if err != nil {
		t.Errorf("Verify failed: %v", err)
	}

	// Test an invalid Dockerfile
	dockerfile = `FROM invalid:latest
COPY . .
RUN invalid-command
CMD ["./app"]`

	err = Verify(context.Background(), fsys, dockerfile)
	if err == nil {
		t.Error("expected verification to fail")
	}
}
