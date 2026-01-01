package logger

import (
	"context"
	"fmt"
	"time"
)

// Logger defines the interface for structured logging.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
	WithFields(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// logger is the concrete implementation of Logger.
type logger struct {
	config  Config
	fields  []Field
	context context.Context
}

// New creates a new logger with the given configuration.
func New(cfg Config) (Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &logger{
		config:  cfg,
		fields:  cfg.Fields,
		context: context.Background(),
	}, nil
}

// Debug logs a debug message.
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	if !l.config.Level.Enabled(DebugLevel) {
		return
	}
	l.log(ctx, DebugLevel, msg, nil, fields...)
}

// Info logs an info message.
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	if !l.config.Level.Enabled(InfoLevel) {
		return
	}
	l.log(ctx, InfoLevel, msg, nil, fields...)
}

// Warn logs a warning message.
func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	if !l.config.Level.Enabled(WarnLevel) {
		return
	}
	l.log(ctx, WarnLevel, msg, nil, fields...)
}

// Error logs an error message.
func (l *logger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	if !l.config.Level.Enabled(ErrorLevel) {
		return
	}
	l.log(ctx, ErrorLevel, msg, err, fields...)
}

// WithFields creates a child logger with additional persistent fields.
func (l *logger) WithFields(fields ...Field) Logger {
	return &logger{
		config:  l.config,
		fields:  append(l.fields, fields...),
		context: l.context,
	}
}

// WithContext creates a child logger with a persistent context.
func (l *logger) WithContext(ctx context.Context) Logger {
	return &logger{
		config:  l.config,
		fields:  l.fields,
		context: ctx,
	}
}

// log writes a log entry.
func (l *logger) log(ctx context.Context, level Level, msg string, err error, fields ...Field) {
	// Merge contexts - use provided ctx if available, otherwise use logger's context
	logCtx := ctx
	if logCtx == nil {
		logCtx = l.context
	}
	if logCtx == nil {
		logCtx = context.Background()
	}

	// Build fields map
	fieldMap := make(map[string]interface{})

	// Add default fields
	for _, field := range l.fields {
		fieldMap[field.Key] = field.Value
	}

	// Add provided fields
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}

	// Add error if provided
	if err != nil {
		fieldMap["error"] = err.Error()
		// Add wrapped error information if available
		if wrapped := unwrapError(err); wrapped != nil {
			fieldMap["error_cause"] = wrapped.Error()
		}
	}

	// Create log entry
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fieldMap,
	}

	// Add caller information if enabled
	if l.config.AddCaller {
		entry.Caller = GetCaller(4) // Skip: log -> Debug/Info/Warn/Error -> logger.log -> GetCaller
	}

	// Add stacktrace for errors if enabled
	if level == ErrorLevel && l.config.AddStacktrace {
		entry.Stacktrace = GetStacktrace()
	}

	// Extract trace context if enabled
	if l.config.EnableTraceCorrelation {
		traceID, spanID := extractTraceContext(logCtx)
		entry.TraceID = traceID
		entry.SpanID = spanID
	}

	// Extract request ID from context if available
	if requestID := extractRequestID(logCtx); requestID != "" {
		entry.RequestID = requestID
	}

	// Write the entry
	_ = l.config.Output.Write(entry)
}

// extractTraceContext extracts trace ID and span ID from context.
// This is a placeholder that can be extended with OpenTelemetry integration.
func extractTraceContext(ctx context.Context) (traceID, spanID string) {
	// Try to extract from context values
	if traceIDVal := ctx.Value("trace_id"); traceIDVal != nil {
		if id, ok := traceIDVal.(string); ok {
			traceID = id
		}
	}
	if spanIDVal := ctx.Value("span_id"); spanIDVal != nil {
		if id, ok := spanIDVal.(string); ok {
			spanID = id
		}
	}
	return traceID, spanID
}

// extractRequestID extracts request ID from context.
func extractRequestID(ctx context.Context) string {
	if requestIDVal := ctx.Value("request_id"); requestIDVal != nil {
		if id, ok := requestIDVal.(string); ok {
			return id
		}
	}
	return ""
}

// unwrapError attempts to unwrap an error to get the underlying cause.
func unwrapError(err error) error {
	type unwrapper interface {
		Unwrap() error
	}
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}
