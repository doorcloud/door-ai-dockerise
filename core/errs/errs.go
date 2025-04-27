package errs

import "fmt"

// Wrap adds context to an error by prefixing it with an operation name.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
