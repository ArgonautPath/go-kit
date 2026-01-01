package logger

import (
	"fmt"
	"time"
)

// Config holds configuration for the logger.
type Config struct {
	// Level is the minimum log level to output.
	Level Level
	// Output is the writer to write logs to.
	Output Writer
	// Format is the output format (JSON or Text).
	Format Format
	// AddCaller includes caller information (file:line) in logs.
	AddCaller bool
	// AddStacktrace includes stack traces for error level logs.
	AddStacktrace bool
	// Fields are default fields to include in all logs.
	Fields []Field
	// TimestampFormat is the format for timestamps.
	TimestampFormat string
	// EnableTraceCorrelation enables OpenTelemetry trace ID injection.
	EnableTraceCorrelation bool
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		Level:            InfoLevel,
		Output:           NewStdoutWriter(JSONFormat),
		Format:           JSONFormat,
		AddCaller:        false,
		AddStacktrace:    false,
		Fields:           nil,
		TimestampFormat:  time.RFC3339,
		EnableTraceCorrelation: false,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Output == nil {
		return fmt.Errorf("output writer is required")
	}
	return nil
}

