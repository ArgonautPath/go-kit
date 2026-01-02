package main

import (
	"context"
	"go-kit/pkg/logger"
)

func main() {
	log, _ := logger.New(logger.Config{
		Level:                  logger.InfoLevel,
		Output:                 logger.NewStdoutWriter(logger.JSONFormat),
		Format:                 logger.JSONFormat,
		EnableTraceCorrelation: true,
		AddCaller:              true,
	})

	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	httpLogger := log.Prefix("[HTTP]")
	httpLogger.Info(ctx, "request processed",
		logger.String("method", "GET"),
		logger.Int("status", 200),
	)
}
