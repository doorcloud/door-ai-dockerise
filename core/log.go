package core

// Logger defines a simple logging interface
type Logger interface {
	Printf(format string, v ...any)
}
