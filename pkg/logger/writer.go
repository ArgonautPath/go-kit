package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// Format represents the output format for logs.
type Format int

const (
	// JSONFormat outputs logs as JSON.
	JSONFormat Format = iota
	// TextFormat outputs logs as human-readable text.
	TextFormat
)

// Writer defines the interface for writing log entries.
type Writer interface {
	Write(entry *LogEntry) error
}

// LogEntry represents a single log entry.
type LogEntry struct {
	Timestamp  time.Time
	Level      Level
	Message    string
	Fields     map[string]interface{}
	Caller     string
	Stacktrace string
	TraceID    string
	SpanID     string
	RequestID  string
}

// stdoutWriter writes logs to os.Stdout.
type stdoutWriter struct {
	format Format
	writer io.Writer
}

// NewStdoutWriter creates a new stdout writer.
func NewStdoutWriter(format Format) Writer {
	return &stdoutWriter{
		format: format,
		writer: os.Stdout,
	}
}

// Write implements the Writer interface.
func (w *stdoutWriter) Write(entry *LogEntry) error {
	return formatEntry(w.writer, w.format, entry)
}

// stderrWriter writes logs to os.Stderr.
type stderrWriter struct {
	format Format
	writer io.Writer
}

// NewStderrWriter creates a new stderr writer.
func NewStderrWriter(format Format) Writer {
	return &stderrWriter{
		format: format,
		writer: os.Stderr,
	}
}

// Write implements the Writer interface.
func (w *stderrWriter) Write(entry *LogEntry) error {
	return formatEntry(w.writer, w.format, entry)
}

// fileWriter writes logs to a file.
type fileWriter struct {
	format Format
	file   *os.File
}

// NewFileWriter creates a new file writer.
func NewFileWriter(path string, format Format) (Writer, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &fileWriter{
		format: format,
		file:   file,
	}, nil
}

// Write implements the Writer interface.
func (w *fileWriter) Write(entry *LogEntry) error {
	return formatEntry(w.file, w.format, entry)
}

// Close closes the file writer.
func (w *fileWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// multiWriter writes logs to multiple writers.
type multiWriter struct {
	writers []Writer
}

// NewMultiWriter creates a new multi writer that writes to all provided writers.
func NewMultiWriter(writers ...Writer) Writer {
	return &multiWriter{
		writers: writers,
	}
}

// Write implements the Writer interface.
func (w *multiWriter) Write(entry *LogEntry) error {
	for _, writer := range w.writers {
		if err := writer.Write(entry); err != nil {
			return err
		}
	}
	return nil
}

// formatEntry formats and writes a log entry.
func formatEntry(w io.Writer, format Format, entry *LogEntry) error {
	switch format {
	case JSONFormat:
		return writeJSON(w, entry)
	case TextFormat:
		return writeText(w, entry)
	default:
		return writeJSON(w, entry)
	}
}

// writeJSON writes a log entry in JSON format.
func writeJSON(w io.Writer, entry *LogEntry) error {
	data := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(time.RFC3339),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}

	// Add fields
	for k, v := range entry.Fields {
		data[k] = v
	}

	// Add optional fields
	if entry.Caller != "" {
		data["caller"] = entry.Caller
	}
	if entry.Stacktrace != "" {
		data["stacktrace"] = entry.Stacktrace
	}
	if entry.TraceID != "" {
		data["trace_id"] = entry.TraceID
	}
	if entry.SpanID != "" {
		data["span_id"] = entry.SpanID
	}
	if entry.RequestID != "" {
		data["request_id"] = entry.RequestID
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encode log entry: %w", err)
	}
	return nil
}

// writeText writes a log entry in human-readable text format.
func writeText(w io.Writer, entry *LogEntry) error {
	var parts []string

	// Timestamp
	parts = append(parts, entry.Timestamp.Format(time.RFC3339))

	// Level
	parts = append(parts, strings.ToUpper(entry.Level.String()))

	// Message
	parts = append(parts, entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		var fieldParts []string
		for k, v := range entry.Fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	// Optional fields
	if entry.Caller != "" {
		parts = append(parts, fmt.Sprintf("caller=%s", entry.Caller))
	}
	if entry.TraceID != "" {
		parts = append(parts, fmt.Sprintf("trace_id=%s", entry.TraceID))
	}
	if entry.SpanID != "" {
		parts = append(parts, fmt.Sprintf("span_id=%s", entry.SpanID))
	}
	if entry.RequestID != "" {
		parts = append(parts, fmt.Sprintf("request_id=%s", entry.RequestID))
	}

	line := strings.Join(parts, " ") + "\n"
	_, err := w.Write([]byte(line))
	return err
}

// GetCaller returns the caller information in the format "file:line".
func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]
	return fmt.Sprintf("%s:%d", filename, line)
}

// GetStacktrace returns the stack trace as a string.
func GetStacktrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
