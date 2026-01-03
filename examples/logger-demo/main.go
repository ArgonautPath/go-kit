package main

import (
	"context"
	"errors"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/logger"
)

func main() {
	// Example 1: Basic JSON logger
	jsonLog, _ := logger.New(logger.Config{
		Level:                  logger.InfoLevel,
		Output:                 logger.NewStdoutWriter(logger.JSONFormat),
		Format:                 logger.JSONFormat,
		EnableTraceCorrelation: true,
		AddCaller:              true,
	})

	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	jsonLog.Info(ctx, "Application started",
		logger.String("version", "1.0.0"),
		logger.String("environment", "production"),
	)

	// Example 2: Text format logger
	textLog, _ := logger.New(logger.Config{
		Level:     logger.DebugLevel,
		Output:    logger.NewStdoutWriter(logger.TextFormat),
		Format:    logger.TextFormat,
		AddCaller: true,
	})

	textLog.Debug(ctx, "Debug message with fields",
		logger.String("component", "database"),
		logger.Duration("query_time", time.Millisecond*45),
	)

	// Example 3: Logger with prefix
	httpLogger := jsonLog.Prefix("[HTTP]")
	httpLogger.Info(ctx, "request processed",
		logger.String("method", "GET"),
		logger.String("path", "/api/users"),
		logger.Int("status", 200),
		logger.Duration("duration", time.Millisecond*150),
	)

	// Example 4: Logger with persistent fields
	dbLogger := jsonLog.WithFields(
		logger.String("service", "database"),
		logger.String("host", "db.example.com"),
	)
	dbLogger.Info(ctx, "Connection established",
		logger.Int("pool_size", 10),
	)

	// Example 5: Error logging
	err := errors.New("failed to connect to database")
	jsonLog.Error(ctx, "Database connection failed", err,
		logger.String("host", "db.example.com"),
		logger.Int("port", 5432),
		logger.Int("retry_count", 3),
	)

	// Example 6: Warning
	jsonLog.Warn(ctx, "High memory usage detected",
		logger.Float64("memory_usage_percent", 85.5),
		logger.String("action", "consider scaling"),
	)

	// Example 7: Async logger (non-blocking)
	asyncLog, _ := logger.New(logger.Config{
		Level:           logger.InfoLevel,
		Output:          logger.NewStdoutWriter(logger.JSONFormat),
		Format:          logger.JSONFormat,
		AsyncEnabled:    true,
		AsyncBufferSize: 1000,
		AddCaller:       true,
	})

	// These log calls return immediately, never blocking
	for i := 0; i < 10; i++ {
		asyncLog.Info(ctx, "High-throughput logging",
			logger.Int("iteration", i),
			logger.String("message", "This is non-blocking"),
		)
	}

	// Always close async logger to flush buffered entries
	defer asyncLog.Close()
}
