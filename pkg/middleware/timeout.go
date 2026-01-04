package middleware

import (
	"context"
	"net/http"
	"time"
)

// TimeoutConfig holds configuration for the Timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration for request handling.
	Timeout time.Duration
	// Message is the error message to return on timeout.
	Message string
	// StatusCode is the HTTP status code to return on timeout.
	StatusCode int
}

// Timeout adds a timeout to request handling. If the handler takes longer
// than the specified timeout, the request is cancelled and an error response
// is returned.
//
// Example:
//
//	mux := http.NewServeMux()
//	handler := Timeout(30 * time.Second)(mux)
func Timeout(timeout time.Duration, opts ...TimeoutOption) Middleware {
	cfg := TimeoutConfig{
		Timeout:    timeout,
		Message:    "Request timeout",
		StatusCode: http.StatusRequestTimeout,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan bool, 1)

			// Create a response writer wrapper
			rw := &timeoutResponseWriter{
				ResponseWriter: w,
				done:           done,
			}

			// Execute handler in goroutine
			go func() {
				next.ServeHTTP(rw, r.WithContext(ctx))
				done <- true
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Timeout occurred
				if !rw.wroteHeader {
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.WriteHeader(cfg.StatusCode)
					w.Write([]byte(cfg.Message))
				}
			}
		})
	}
}

// TimeoutOption is a functional option for Timeout middleware.
type TimeoutOption func(*TimeoutConfig)

// WithTimeoutMessage sets the error message for timeout responses.
func WithTimeoutMessage(message string) TimeoutOption {
	return func(cfg *TimeoutConfig) {
		cfg.Message = message
	}
}

// WithTimeoutStatusCode sets the HTTP status code for timeout responses.
func WithTimeoutStatusCode(code int) TimeoutOption {
	return func(cfg *TimeoutConfig) {
		cfg.StatusCode = code
	}
}

// timeoutResponseWriter wraps http.ResponseWriter to track if headers were written.
type timeoutResponseWriter struct {
	http.ResponseWriter
	done        chan bool
	wroteHeader bool
}

func (rw *timeoutResponseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *timeoutResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

