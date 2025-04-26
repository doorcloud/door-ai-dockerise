package plain

import (
	"fmt"
	"io"
	"os"
)

// Streamer implements core.LogStreamer with colorized output
type Streamer struct {
	out io.Writer
}

// New creates a new plain text log streamer
func New() *Streamer {
	return &Streamer{out: os.Stdout}
}

// Info writes an informational message in blue
func (s *Streamer) Info(msg string) {
	fmt.Fprintf(s.out, "\033[34m%s\033[0m\n", msg)
}

// Error writes an error message in red
func (s *Streamer) Error(msg string) {
	fmt.Fprintf(s.out, "\033[31m%s\033[0m\n", msg)
}

// Write implements io.Writer for raw log output
func (s *Streamer) Write(p []byte) (n int, err error) {
	return s.out.Write(p)
}
