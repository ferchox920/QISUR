package logger

import "log"

// New returns a basic logger; replace with structured logging as needed.
func New() *log.Logger {
	return log.Default()
}
