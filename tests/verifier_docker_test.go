package tests

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/doorcloud/door-ai-dockerise/adapters/facts/spring"
	"github.com/doorcloud/door-ai-dockerise/adapters/verifiers/docker"
	"github.com/doorcloud/door-ai-dockerise/drivers/docker/mock"
)

func TestDockerVerifier(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Parse server URL to get port
		serverURL, err := url.Parse(server.URL)
		if err != nil {
			t.Fatalf("failed to parse server URL: %v", err)
		}
		port, err := strconv.Atoi(serverURL.Port())
		if err != nil {
			t.Fatalf("failed to parse server port: %v", err)
		}

		// Create mock Docker client
		mockClient := mock.NewMockDriver()
		mockClient.GetRandomPortFunc = func() (int, error) {
			return port, nil
		}
		mockClient.RunFunc = func(ctx context.Context, image string, hostPort, containerPort int) (string, error) {
			return "test-container-id", nil
		}
		mockClient.LogsFunc = func(ctx context.Context, containerID string, w io.Writer) error {
			return nil
		}

		// Create verifier
		verifier := docker.NewRunner(mockClient)

		// Create test spec
		spec := &spring.Spec{
			Ports:          []int{8080},
			HealthEndpoint: "/health",
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create buffer for logs
		var logs strings.Builder

		// Run verifier
		err = verifier.Run(ctx, "test-image", spec, &logs)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Verify logs
		expectedLogs := "docker run   │ test-container-id\nhealth       │ OK in"
		if !strings.Contains(logs.String(), expectedLogs) {
			t.Errorf("logs do not contain expected output. Got: %s", logs.String())
		}
	})

	t.Run("timeout", func(t *testing.T) {
		// Create mock Docker client
		mockClient := mock.NewMockDriver()
		mockClient.GetRandomPortFunc = func() (int, error) {
			return 8080, nil
		}
		mockClient.RunFunc = func(ctx context.Context, image string, hostPort, containerPort int) (string, error) {
			return "test-container-id", nil
		}
		mockClient.LogsFunc = func(ctx context.Context, containerID string, w io.Writer) error {
			return nil
		}

		// Create verifier
		verifier := docker.NewRunner(mockClient)

		// Create test spec
		spec := &spring.Spec{
			Ports:          []int{8080},
			HealthEndpoint: "/health",
		}

		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Create buffer for logs
		var logs strings.Builder

		// Run verifier
		err := verifier.Run(ctx, "test-image", spec, &logs)
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected DeadlineExceeded error, got: %v", err)
		}
	})
}
