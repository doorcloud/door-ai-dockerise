package json

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Streamer implements core.LogStreamer with JSON output
type Streamer struct {
	out io.Writer
	enc *json.Encoder
}

// logEntry represents a single log entry
type logEntry struct {
	Timestamp time.Time `json:"ts"`
	Level     string    `json:"level"`
	Message   string    `json:"msg"`
}

// New creates a new JSON log streamer
func New() *Streamer {
	s := &Streamer{out: os.Stdout}
	s.enc = json.NewEncoder(s.out)
	return s
}

// Info writes an informational message
func (s *Streamer) Info(msg string) {
	s.write("info", msg)
}

// Error writes an error message
func (s *Streamer) Error(msg string) {
	s.write("error", msg)
}

// Write implements io.Writer for raw log output
func (s *Streamer) Write(p []byte) (n int, err error) {
	s.write("raw", string(p))
	return len(p), nil
}

func (s *Streamer) write(level, msg string) {
	entry := logEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
	}
	s.enc.Encode(entry)
}
