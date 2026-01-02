package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// mockWriter is a mock writer for testing.
type mockWriter struct {
	entries []*LogEntry
	buf     *bytes.Buffer
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		entries: make([]*LogEntry, 0),
		buf:     &bytes.Buffer{},
	}
}

func (m *mockWriter) Write(entry *LogEntry) error {
	m.entries = append(m.entries, entry)
	return formatEntry(m.buf, JSONFormat, entry)
}

func (m *mockWriter) String() string {
	return m.buf.String()
}

func (m *mockWriter) Reset() {
	m.entries = m.entries[:0]
	m.buf.Reset()
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{"debug", DebugLevel, "debug"},
		{"info", InfoLevel, "info"},
		{"warn", WarnLevel, "warn"},
		{"error", ErrorLevel, "error"},
		{"unknown", Level(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{"debug", "debug", DebugLevel, false},
		{"info", "info", InfoLevel, false},
		{"warn", "warn", WarnLevel, false},
		{"warning", "warning", WarnLevel, false},
		{"error", "error", ErrorLevel, false},
		{"uppercase", "DEBUG", DebugLevel, false},
		{"mixed", "Info", InfoLevel, false},
		{"invalid", "invalid", DebugLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLevel_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		minLevel Level
		level    Level
		want     bool
	}{
		{"debug enabled at debug", DebugLevel, DebugLevel, true},
		{"info enabled at debug", DebugLevel, InfoLevel, true},
		{"warn enabled at debug", DebugLevel, WarnLevel, true},
		{"error enabled at debug", DebugLevel, ErrorLevel, true},
		{"debug disabled at info", InfoLevel, DebugLevel, false},
		{"info enabled at info", InfoLevel, InfoLevel, true},
		{"warn enabled at info", InfoLevel, WarnLevel, true},
		{"error enabled at info", InfoLevel, ErrorLevel, true},
		{"debug disabled at warn", WarnLevel, DebugLevel, false},
		{"info disabled at warn", WarnLevel, InfoLevel, false},
		{"warn enabled at warn", WarnLevel, WarnLevel, true},
		{"error enabled at warn", WarnLevel, ErrorLevel, true},
		{"debug disabled at error", ErrorLevel, DebugLevel, false},
		{"info disabled at error", ErrorLevel, InfoLevel, false},
		{"warn disabled at error", ErrorLevel, WarnLevel, false},
		{"error enabled at error", ErrorLevel, ErrorLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.minLevel.Enabled(tt.level); got != tt.want {
				t.Errorf("Level.Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		key   string
		value interface{}
	}{
		{"string", String("key", "value"), "key", "value"},
		{"int", Int("key", 42), "key", 42},
		{"int64", Int64("key", 64), "key", int64(64)},
		{"float64", Float64("key", 3.14), "key", 3.14},
		{"bool", Bool("key", true), "key", true},
		{"error", Error(errors.New("test error")), "error", "test error"},
		{"any", Any("key", "any value"), "key", "any value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Key != tt.key {
				t.Errorf("Field.Key = %v, want %v", tt.field.Key, tt.key)
			}
			if tt.field.Value != tt.value {
				t.Errorf("Field.Value = %v, want %v", tt.field.Value, tt.value)
			}
		})
	}
}

func TestFields_MapConversion(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	fields := Fields(m)
	if len(fields) != 3 {
		t.Errorf("Fields() length = %v, want 3", len(fields))
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Level:  InfoLevel,
				Output: newMockWriter(),
				Format: JSONFormat,
			},
			wantErr: false,
		},
		{
			name: "missing output",
			config: Config{
				Level:  InfoLevel,
				Output: nil,
				Format: JSONFormat,
			},
			wantErr: true,
		},
		{
			name:    "default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger without error")
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  DebugLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	logger.Debug(ctx, "debug message", String("key", "value"))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Level != DebugLevel {
		t.Errorf("Entry.Level = %v, want %v", entry.Level, DebugLevel)
	}
	if entry.Message != "debug message" {
		t.Errorf("Entry.Message = %v, want %v", entry.Message, "debug message")
	}
	if entry.Fields["key"] != "value" {
		t.Errorf("Entry.Fields[key] = %v, want %v", entry.Fields["key"], "value")
	}
}

func TestLogger_Info(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  InfoLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	logger.Info(ctx, "info message", Int("count", 42))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Level != InfoLevel {
		t.Errorf("Entry.Level = %v, want %v", entry.Level, InfoLevel)
	}
	if entry.Message != "info message" {
		t.Errorf("Entry.Message = %v, want %v", entry.Message, "info message")
	}
	if entry.Fields["count"] != 42 {
		t.Errorf("Entry.Fields[count] = %v, want %v", entry.Fields["count"], 42)
	}
}

func TestLogger_Warn(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  WarnLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	logger.Warn(ctx, "warn message", Bool("flag", true))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Level != WarnLevel {
		t.Errorf("Entry.Level = %v, want %v", entry.Level, WarnLevel)
	}
}

func TestLogger_Error(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  ErrorLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	testErr := errors.New("test error")
	logger.Error(ctx, "error message", testErr, String("component", "test"))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Level != ErrorLevel {
		t.Errorf("Entry.Level = %v, want %v", entry.Level, ErrorLevel)
	}
	if entry.Fields["error"] != "test error" {
		t.Errorf("Entry.Fields[error] = %v, want %v", entry.Fields["error"], "test error")
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		minLevel  Level
		logLevel  Level
		logFunc   func(Logger, context.Context, string, ...Field)
		shouldLog bool
	}{
		{"debug at debug", DebugLevel, DebugLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Debug(ctx, msg, fields...) }, true},
		{"debug at info", InfoLevel, DebugLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Debug(ctx, msg, fields...) }, false},
		{"info at info", InfoLevel, InfoLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Info(ctx, msg, fields...) }, true},
		{"info at warn", WarnLevel, InfoLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Info(ctx, msg, fields...) }, false},
		{"warn at warn", WarnLevel, WarnLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Warn(ctx, msg, fields...) }, true},
		{"warn at error", ErrorLevel, WarnLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Warn(ctx, msg, fields...) }, false},
		{"error at error", ErrorLevel, ErrorLevel, func(l Logger, ctx context.Context, msg string, fields ...Field) { l.Error(ctx, msg, nil, fields...) }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockWriter()
			logger, _ := New(Config{
				Level:  tt.minLevel,
				Output: mock,
				Format: JSONFormat,
			})

			ctx := context.Background()
			tt.logFunc(logger, ctx, "test message")

			if tt.shouldLog && len(mock.entries) != 1 {
				t.Errorf("Expected 1 entry, got %d", len(mock.entries))
			}
			if !tt.shouldLog && len(mock.entries) != 0 {
				t.Errorf("Expected 0 entries, got %d", len(mock.entries))
			}
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  InfoLevel,
		Output: mock,
		Format: JSONFormat,
	})

	childLogger := logger.WithFields(String("service", "test"), Int("version", 1))
	ctx := context.Background()
	childLogger.Info(ctx, "message", String("key", "value"))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Fields["service"] != "test" {
		t.Errorf("Entry.Fields[service] = %v, want %v", entry.Fields["service"], "test")
	}
	if entry.Fields["version"] != 1 {
		t.Errorf("Entry.Fields[version] = %v, want %v", entry.Fields["version"], 1)
	}
	if entry.Fields["key"] != "value" {
		t.Errorf("Entry.Fields[key] = %v, want %v", entry.Fields["key"], "value")
	}
}

func TestLogger_WithContext(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:                  InfoLevel,
		Output:                 mock,
		Format:                 JSONFormat,
		EnableTraceCorrelation: true,
	})

	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	ctx = context.WithValue(ctx, "span_id", "span-456")
	ctx = context.WithValue(ctx, "request_id", "req-789")

	childLogger := logger.WithContext(ctx)
	childLogger.Info(ctx, "message")

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.TraceID != "trace-123" {
		t.Errorf("Entry.TraceID = %v, want %v", entry.TraceID, "trace-123")
	}
	if entry.SpanID != "span-456" {
		t.Errorf("Entry.SpanID = %v, want %v", entry.SpanID, "span-456")
	}
	if entry.RequestID != "req-789" {
		t.Errorf("Entry.RequestID = %v, want %v", entry.RequestID, "req-789")
	}
}

func TestLogger_AddCaller(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:     InfoLevel,
		Output:    mock,
		Format:    JSONFormat,
		AddCaller: true,
	})

	ctx := context.Background()
	logger.Info(ctx, "message")

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Caller == "" {
		t.Error("Entry.Caller should not be empty when AddCaller is enabled")
	}
	if !strings.Contains(entry.Caller, ".go:") {
		t.Errorf("Entry.Caller = %v, should contain file:line format", entry.Caller)
	}
}

func TestLogger_AddStacktrace(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:         ErrorLevel,
		Output:        mock,
		Format:        JSONFormat,
		AddStacktrace: true,
	})

	ctx := context.Background()
	logger.Error(ctx, "error message", errors.New("test error"))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	if entry.Stacktrace == "" {
		t.Error("Entry.Stacktrace should not be empty when AddStacktrace is enabled")
	}
}

func TestLogger_Prefix(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  InfoLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	prefixedLogger := logger.Prefix("[HTTP]")
	prefixedLogger.Info(ctx, "request received")

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	expectedMsg := "[HTTP] request received"
	if entry.Message != expectedMsg {
		t.Errorf("Entry.Message = %q, want %q", entry.Message, expectedMsg)
	}
}

func TestLogger_PrefixChaining(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  InfoLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	// Chain multiple prefixes
	prefixedLogger := logger.Prefix("[HTTP]").Prefix("[Handler]")
	prefixedLogger.Info(ctx, "processing request")

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	expectedMsg := "[HTTP] [Handler] processing request"
	if entry.Message != expectedMsg {
		t.Errorf("Entry.Message = %q, want %q", entry.Message, expectedMsg)
	}
}

func TestLogger_PrefixWithFields(t *testing.T) {
	mock := newMockWriter()
	logger, _ := New(Config{
		Level:  InfoLevel,
		Output: mock,
		Format: JSONFormat,
	})

	ctx := context.Background()
	// Combine prefix with fields
	prefixedLogger := logger.Prefix("[DB]").WithFields(String("table", "users"))
	prefixedLogger.Info(ctx, "query executed", Int("rows", 10))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	expectedMsg := "[DB] query executed"
	if entry.Message != expectedMsg {
		t.Errorf("Entry.Message = %q, want %q", entry.Message, expectedMsg)
	}
	if entry.Fields["table"] != "users" {
		t.Errorf("Entry.Fields[table] = %v, want %v", entry.Fields["table"], "users")
	}
	if entry.Fields["rows"] != 10 {
		t.Errorf("Entry.Fields[rows] = %v, want %v", entry.Fields["rows"], 10)
	}
}

func TestWriter_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := &mockWriter{buf: buf}

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "test message",
		Fields: map[string]interface{}{
			"key": "value",
		},
	}

	err := writer.Write(entry)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if data["level"] != "info" {
		t.Errorf("data[level] = %v, want %v", data["level"], "info")
	}
	if data["message"] != "test message" {
		t.Errorf("data[message] = %v, want %v", data["message"], "test message")
	}
	if data["key"] != "value" {
		t.Errorf("data[key] = %v, want %v", data["key"], "value")
	}
}

func TestWriter_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := &textWriter{format: TextFormat, writer: buf}

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "test message",
		Fields: map[string]interface{}{
			"key": "value",
		},
	}

	err := writer.Write(entry)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Output should contain 'INFO', got: %v", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Output should contain 'test message', got: %v", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Output should contain 'key=value', got: %v", output)
	}
}

// textWriter is a helper for testing text format
type textWriter struct {
	format Format
	writer io.Writer
}

func (w *textWriter) Write(entry *LogEntry) error {
	return formatEntry(w.writer, w.format, entry)
}

func TestMultiWriter(t *testing.T) {
	mock1 := newMockWriter()
	mock2 := newMockWriter()
	multi := NewMultiWriter(mock1, mock2)

	entry := &LogEntry{
		Level:   InfoLevel,
		Message: "test",
		Fields:  make(map[string]interface{}),
	}

	err := multi.Write(entry)
	if err != nil {
		t.Fatalf("MultiWriter.Write() error = %v", err)
	}

	if len(mock1.entries) != 1 {
		t.Errorf("mock1 should have 1 entry, got %d", len(mock1.entries))
	}
	if len(mock2.entries) != 1 {
		t.Errorf("mock2 should have 1 entry, got %d", len(mock2.entries))
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Level != InfoLevel {
		t.Errorf("DefaultConfig().Level = %v, want %v", cfg.Level, InfoLevel)
	}
	if cfg.Output == nil {
		t.Error("DefaultConfig().Output should not be nil")
	}
	if cfg.Format != JSONFormat {
		t.Errorf("DefaultConfig().Format = %v, want %v", cfg.Format, JSONFormat)
	}
}

func TestAsyncLogger_NonBlocking(t *testing.T) {
	mock := newMockWriter()
	logger, err := New(Config{
		Level:           InfoLevel,
		Output:          mock,
		Format:          JSONFormat,
		AsyncEnabled:    true,
		AsyncBufferSize: 10,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	ctx := context.Background()

	// Write multiple entries rapidly - should not block
	for i := 0; i < 20; i++ {
		logger.Info(ctx, "test message", Int("index", i))
	}

	// Give the async worker time to process
	time.Sleep(100 * time.Millisecond)

	// Should have processed at least some entries (up to buffer size)
	if len(mock.entries) == 0 {
		t.Error("Expected some entries to be written")
	}
}

func TestAsyncLogger_DropWhenFull(t *testing.T) {
	mock := newMockWriter()
	logger, err := New(Config{
		Level:           InfoLevel,
		Output:          mock,
		Format:          JSONFormat,
		AsyncEnabled:    true,
		AsyncBufferSize: 5, // Small buffer
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Close()

	ctx := context.Background()

	// Fill buffer and overflow
	for i := 0; i < 100; i++ {
		logger.Info(ctx, "test message", Int("index", i))
	}

	// Give the async worker time to process
	time.Sleep(200 * time.Millisecond)

	// Should have written some entries, but not all due to drops
	// The exact number depends on timing, but should be less than 100
	if len(mock.entries) >= 100 {
		t.Errorf("Expected some entries to be dropped, but got %d entries", len(mock.entries))
	}
}

func TestAsyncLogger_GracefulShutdown(t *testing.T) {
	mock := newMockWriter()
	logger, err := New(Config{
		Level:           InfoLevel,
		Output:          mock,
		Format:          JSONFormat,
		AsyncEnabled:    true,
		AsyncBufferSize: 10,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()

	// Write some entries
	for i := 0; i < 5; i++ {
		logger.Info(ctx, "test message", Int("index", i))
	}

	// Close should drain and wait for all entries to be written
	err = logger.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// All entries should be written after close
	if len(mock.entries) != 5 {
		t.Errorf("Expected 5 entries after close, got %d", len(mock.entries))
	}

	// Closing again should be safe
	err = logger.Close()
	if err != nil {
		t.Errorf("Close() second time error = %v", err)
	}
}

func TestAsyncLogger_ChildLoggers(t *testing.T) {
	mock := newMockWriter()
	baseLogger, err := New(Config{
		Level:           InfoLevel,
		Output:          mock,
		Format:          JSONFormat,
		AsyncEnabled:    true,
		AsyncBufferSize: 10,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer baseLogger.Close()

	ctx := context.Background()

	// Create child loggers
	child1 := baseLogger.WithFields(String("source", "child1"))
	child2 := baseLogger.Prefix("[PREFIX]")

	// Write from different loggers
	child1.Info(ctx, "message 1")
	child2.Info(ctx, "message 2")
	baseLogger.Info(ctx, "message 3")

	// Give time to process
	time.Sleep(100 * time.Millisecond)

	// All should be written
	if len(mock.entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(mock.entries))
	}
}
