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
