package plain

import (
	"io"

	"github.com/doorcloud/door-ai-dockerise/core"
)

// WriterStreamer wraps an io.Writer to implement core.LogStreamer
type WriterStreamer struct {
	w io.Writer
}

// NewWriterStreamer creates a new LogStreamer that wraps an io.Writer
func NewWriterStreamer(w io.Writer) core.LogStreamer {
	return &WriterStreamer{w: w}
}

// Info writes an informational message
func (s *WriterStreamer) Info(msg string) {
	s.w.Write([]byte(msg + "\n"))
}

// Error writes an error message
func (s *WriterStreamer) Error(msg string) {
	s.w.Write([]byte("ERROR: " + msg + "\n"))
}

// Write implements io.Writer
func (s *WriterStreamer) Write(p []byte) (n int, err error) {
	return s.w.Write(p)
}
