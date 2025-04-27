package logs

import "log"

// Tag prints aligned "[tag] │ ..." lines for streaming logs.
func Tag(tag string, format string, a ...any) {
	log.Printf("%-10s │ "+format, append([]any{tag}, a...)...)
}
