package e2e

import (
	"bufio"
	"bytes"
	"io"
	"testing"
	"time"
)

// AssertLog scans the reader's output until the given substring is found
// or the timeout is reached. Fails the test if the substring is not found.
func AssertLog(t *testing.T, r io.Reader, substr string, d time.Duration) {
	t.Helper()

	// Create a scanner for the reader
	scanner := bufio.NewScanner(r)

	// Create a channel to signal when the substring is found
	found := make(chan bool)

	// Start scanning in a goroutine
	go func() {
		for scanner.Scan() {
			if scanner.Err() != nil {
				break
			}
			if bytes.Contains(scanner.Bytes(), []byte(substr)) {
				found <- true
				return
			}
		}
		found <- false
	}()

	// Wait for either the substring to be found or timeout
	select {
	case success := <-found:
		if !success {
			t.Errorf("Substring %q not found in logs", substr)
		}
	case <-time.After(d):
		t.Errorf("Timeout waiting for substring %q in logs", substr)
	}
}
