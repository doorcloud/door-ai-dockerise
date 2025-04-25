package dockerverify

import (
	"context"
	"os/exec"
	"testing"
	"testing/fstest"
	"time"
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

	ok, errLog, err := Verify(context.Background(), fsys, dockerfile, 1*time.Minute)
	if err != nil {
		t.Errorf("Verify failed: %v", err)
	}
	if !ok {
		t.Errorf("expected verification to pass, got error log: %s", errLog)
	}

	// Test an invalid Dockerfile
	dockerfile = `FROM invalid:latest
COPY . .
RUN invalid-command
CMD ["./app"]`

	ok, errLog, err = Verify(context.Background(), fsys, dockerfile, 1*time.Minute)
	if err != nil {
		t.Errorf("Verify failed: %v", err)
	}
	if ok {
		t.Error("expected verification to fail")
	}
	if errLog == "" {
		t.Error("expected error log to be non-empty")
	}
}
