package v1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	lines []string
}

func (l *testLogger) Printf(format string, v ...any) {
	l.lines = append(l.lines, fmt.Sprintf(format, v...))
}

func TestOrchestrator_Logging(t *testing.T) {
	logger := &testLogger{}
	o := New(Opts{
		Log: logger,
	})

	// Test logf
	o.logf("Starting build...")
	assert.Len(t, logger.lines, 1)
	assert.Contains(t, logger.lines[0], "Starting build...")
}
