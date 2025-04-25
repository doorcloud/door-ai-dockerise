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

	// Create test filesystem with go.mod
	fsys := fstest.MapFS{
		"go.mod": &fstest.MapFile{
			Data: []byte(`module test
go 1.21

require (
	github.com/example/pkg v1.0.0
)
`),
		},
		"main.go": &fstest.MapFile{
			Data: []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`),
		},
	}

	// Test Dockerfile
	dockerfile := `FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o app
CMD ["./app"]`

	err := Verify(context.Background(), fsys, dockerfile)
	if err != nil {
		t.Errorf("Verify() error = %v", err)
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
