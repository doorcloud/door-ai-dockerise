package core

import (
	"strings"
)

// StringLogSink is a log sink that stores logs in a string
type StringLogSink struct {
	builder strings.Builder
}

// Log implements the LogSink interface
func (s *StringLogSink) Log(msg string) {
	s.builder.WriteString(msg)
	s.builder.WriteString("\n")
}

// String returns the accumulated logs
func (s *StringLogSink) String() string {
	return s.builder.String()
}

// NullLogSink is a log sink that discards all logs
type NullLogSink struct{}

// Log implements the LogSink interface
func (n *NullLogSink) Log(msg string) {
	// Discard the message
}
