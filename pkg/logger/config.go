package logger

import (
	"fmt"
	"time"
)

// Config holds configuration for the logger.
type Config struct {
	// Level is the minimum log level to output. Logs below this level will be filtered out.
	// Example: If set to InfoLevel, Debug logs will be ignored, but Info, Warn, and Error will be logged.
	// Valid values: DebugLevel, InfoLevel, WarnLevel, ErrorLevel
	Level Level

	// Output is the writer to write logs to. This determines where log entries are sent.
	// Examples:
	//   - NewStdoutWriter(JSONFormat) - Write to standard output
	//   - NewStderrWriter(JSONFormat) - Write to standard error
	//   - NewFileWriter("/var/log/app.log", JSONFormat) - Write to a file
	//   - NewMultiWriter(w1, w2) - Write to multiple targets simultaneously
	Output Writer

	// Format is the output format for log entries. Determines how logs are serialized.
	// Examples:
	//   - JSONFormat: {"level":"info","message":"hello","timestamp":"2024-01-01T00:00:00Z"}
	//   - TextFormat: 2024-01-01T00:00:00Z INFO hello
	Format Format

	// AddCaller includes caller information (file:line) in logs. When enabled, each log entry
	// will include the source file and line number where the log was called from.
	// Example output: "caller":"handler.go:42" - useful for debugging to locate log statements.
	// Note: Has minimal performance overhead due to runtime.Caller() call.
	AddCaller bool

	// AddStacktrace includes stack traces for error level logs. When enabled, Error() calls
	// will include a full stack trace showing the call chain leading to the error.
	// Example: Useful for debugging production errors to understand the execution path.
	// Note: Only applies to Error level logs, has performance overhead, use judiciously.
	AddStacktrace bool

	// Fields are default fields to include in all logs. These fields are automatically added
	// to every log entry, providing consistent context across all logs.
	// Example:
	//   Fields: []Field{
	//     String("service", "user-service"),
	//     String("version", "1.2.3"),
	//     String("environment", "production"),
	//   }
	// Every log will automatically include: service=user-service, version=1.2.3, environment=production
	Fields []Field

	// TimestampFormat is the format for timestamps in log entries. Uses Go's time format layout.
	// Example formats:
	//   - time.RFC3339: "2006-01-02T15:04:05Z07:00"
	//   - time.RFC3339Nano: "2006-01-02T15:04:05.999999999Z07:00"
	//   - "2006-01-02 15:04:05": "2024-01-01 12:00:00"
	// Default: time.RFC3339
	TimestampFormat string

	// EnableTraceCorrelation enables OpenTelemetry trace ID and span ID injection into logs.
	// When enabled, the logger will extract trace_id and span_id from the context and include
	// them in log entries, allowing correlation between logs and distributed traces.
	// Example output: "trace_id":"abc123", "span_id":"def456"
	// Note: Requires trace context to be propagated via context.Context. Works with
	// OpenTelemetry or custom trace context values in context.
	EnableTraceCorrelation bool

	// AsyncEnabled enables asynchronous logging. When enabled, log entries are queued in a
	// buffered channel and written by a background goroutine, preventing blocking of the
	// calling goroutine. If the buffer is full, new entries are dropped (non-blocking).
	// Example: Set to true for high-throughput applications where logging should not block.
	// Default: false (synchronous logging)
	AsyncEnabled bool

	// AsyncBufferSize is the size of the buffer channel for async logging. When the buffer
	// is full, new log entries are dropped to prevent blocking. Larger buffers reduce drops
	// but use more memory.
	// Example: 1000 entries - can queue up to 1000 log entries before dropping.
	// Default: 1000
	AsyncBufferSize int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		Level:                  InfoLevel,
		Output:                 NewStdoutWriter(JSONFormat),
		Format:                 JSONFormat,
		AddCaller:              false,
		AddStacktrace:          false,
		Fields:                 nil,
		TimestampFormat:        time.RFC3339,
		EnableTraceCorrelation: false,
		AsyncEnabled:           false,
		AsyncBufferSize:        1000,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Output == nil {
		return fmt.Errorf("output writer is required")
	}
	return nil
}
