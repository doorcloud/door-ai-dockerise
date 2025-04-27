package logs

import (
	"fmt"
	"io"
)

// WriteTagged writes a tagged log line to the given writer
func WriteTagged(w io.Writer, tag string, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(w, "%s │ %s\n", tag, msg)
}

// TagWriter returns a writer that prefixes each line with a tag
type TagWriter struct {
	w   io.Writer
	tag string
}

// NewTagWriter creates a new tagged writer
func NewTagWriter(w io.Writer, tag string) *TagWriter {
	return &TagWriter{w: w, tag: tag}
}

// Write implements io.Writer
func (w *TagWriter) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(w.w, "%s │ %s", w.tag, string(p))
}
