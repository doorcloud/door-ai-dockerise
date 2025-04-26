package mock

import (
	"sync"
)

// Streamer implements core.LogStreamer for testing
type Streamer struct {
	mu      sync.Mutex
	entries []Entry
}

// Entry represents a logged message
type Entry struct {
	Level   string
	Message string
}

// New creates a new mock log streamer
func New() *Streamer {
	return &Streamer{}
}

// Info writes an informational message
func (s *Streamer) Info(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, Entry{Level: "info", Message: msg})
}

// Error writes an error message
func (s *Streamer) Error(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, Entry{Level: "error", Message: msg})
}

// Write implements io.Writer for raw log output
func (s *Streamer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, Entry{Level: "raw", Message: string(p)})
	return len(p), nil
}

// Entries returns all logged entries
func (s *Streamer) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries
}

// Clear clears all logged entries
func (s *Streamer) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = nil
}
