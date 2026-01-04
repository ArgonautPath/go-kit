package middleware

import (
	"net/http"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/logger"
)

// LoggingConfig holds configuration for the Logging middleware.
type LoggingConfig struct {
	// Logger is the logger instance to use. If nil, logging is skipped.
	Logger logger.Logger
	// LogRequestHeaders logs request headers.
	LogRequestHeaders bool
	// LogResponseHeaders logs response headers.
	LogResponseHeaders bool
	// LogRequestBody logs request body (use with caution for large bodies).
	LogRequestBody bool
	// LogResponseBody logs response body (use with caution for large bodies).
	LogResponseBody bool
	// SkipPaths is a list of paths to skip logging.
	SkipPaths []string
	// SkipStatusCodes is a list of HTTP status codes to skip logging.
	SkipStatusCodes []int
}

// Logging logs HTTP requests and responses using the provided logger.
// It logs request method, path, status code, duration, and optionally
// headers and bodies.
//
// Example:
//
//	log, _ := logger.New(logger.Config{...})
//	mux := http.NewServeMux()
//	handler := Logging(log)(mux)
func Logging(l logger.Logger, opts ...LoggingOption) Middleware {
	cfg := LoggingConfig{
		Logger:             l,
		LogRequestHeaders:  false,
		LogResponseHeaders: false,
		LogRequestBody:     false,
		LogResponseBody:    false,
		SkipPaths:          []string{},
		SkipStatusCodes:    []int{},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging if path is in skip list
			if contains(cfg.SkipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip if logger is nil
			if cfg.Logger == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Create response writer wrapper to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Execute next handler
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// Skip logging if status code is in skip list
			if containsInt(cfg.SkipStatusCodes, rw.statusCode) {
				return
			}

			// Build log fields
			ctx := r.Context()
			fields := []logger.Field{
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("query", r.URL.RawQuery),
				logger.Int("status", rw.statusCode),
				logger.Duration("duration", duration),
				logger.String("remote_addr", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
			}

			// Add request ID if available
			if requestID := GetRequestID(ctx); requestID != "" {
				fields = append(fields, logger.String("request_id", requestID))
			}

			// Add request headers if enabled
			if cfg.LogRequestHeaders {
				fields = append(fields, logger.Any("request_headers", r.Header))
			}

			// Add response headers if enabled
			if cfg.LogResponseHeaders {
				fields = append(fields, logger.Any("response_headers", rw.Header()))
			}

			// Log based on status code
			if rw.statusCode >= 500 {
				cfg.Logger.Error(ctx, "HTTP request error", nil, fields...)
			} else if rw.statusCode >= 400 {
				cfg.Logger.Warn(ctx, "HTTP request warning", fields...)
			} else {
				cfg.Logger.Info(ctx, "HTTP request", fields...)
			}
		})
	}
}

// LoggingOption is a functional option for Logging middleware.
type LoggingOption func(*LoggingConfig)

// WithLogRequestHeaders enables logging of request headers.
func WithLogRequestHeaders(enabled bool) LoggingOption {
	return func(cfg *LoggingConfig) {
		cfg.LogRequestHeaders = enabled
	}
}

// WithLogResponseHeaders enables logging of response headers.
func WithLogResponseHeaders(enabled bool) LoggingOption {
	return func(cfg *LoggingConfig) {
		cfg.LogResponseHeaders = enabled
	}
}

// WithLogRequestBody enables logging of request body.
func WithLogRequestBody(enabled bool) LoggingOption {
	return func(cfg *LoggingConfig) {
		cfg.LogRequestBody = enabled
	}
}

// WithLogResponseBody enables logging of response body.
func WithLogResponseBody(enabled bool) LoggingOption {
	return func(cfg *LoggingConfig) {
		cfg.LogResponseBody = enabled
	}
}

// WithSkipPaths sets paths to skip logging.
func WithSkipPaths(paths ...string) LoggingOption {
	return func(cfg *LoggingConfig) {
		cfg.SkipPaths = paths
	}
}

// WithSkipStatusCodes sets status codes to skip logging.
func WithSkipStatusCodes(codes ...int) LoggingOption {
	return func(cfg *LoggingConfig) {
		cfg.SkipStatusCodes = codes
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// containsInt checks if an int slice contains a value.
func containsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

