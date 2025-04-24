package dbg

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Printf logs a debug message if DG_DEBUG is set
func Printf(format string, args ...interface{}) {
	if os.Getenv("DG_DEBUG") != "" {
		log.Printf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	}
}
