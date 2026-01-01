package logger

import (
	"fmt"
	"strings"
)

// Level represents a log level.
type Level int

const (
	// DebugLevel is the lowest level, used for detailed debugging information.
	DebugLevel Level = iota
	// InfoLevel is used for general informational messages.
	InfoLevel
	// WarnLevel is used for warning messages that don't stop execution.
	WarnLevel
	// ErrorLevel is used for error messages that indicate failures.
	ErrorLevel
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel parses a string into a log level.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return DebugLevel, fmt.Errorf("unknown log level: %s", s)
	}
}

// Enabled returns true if the given level is enabled for this level.
// A level is enabled if it is greater than or equal to the current level.
func (l Level) Enabled(level Level) bool {
	return level >= l
}

