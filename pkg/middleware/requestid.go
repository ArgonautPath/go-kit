package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the standard header name for request IDs.
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for storing request IDs.
	RequestIDContextKey contextKey = "request_id"
)

type contextKey string

// RequestIDConfig holds configuration for the RequestID middleware.
type RequestIDConfig struct {
	// HeaderName is the HTTP header name to use for request IDs.
	// Default: "X-Request-ID"
	HeaderName string
	// GenerateID is a function to generate request IDs.
	// If nil, uses UUID v4.
	GenerateID func() string
	// AddToResponse adds the request ID to the response headers.
	// Default: true
	AddToResponse bool
}

// RequestID injects a request ID into the request context and optionally
// adds it to response headers. The request ID is extracted from the request
// header if present, otherwise a new one is generated.
//
// The request ID can be retrieved from the context using GetRequestID.
//
// Example:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", handler)
//	handler := RequestID()(mux)
//	http.ListenAndServe(":8080", handler)
func RequestID(opts ...RequestIDOption) Middleware {
	cfg := RequestIDConfig{
		HeaderName:    RequestIDHeader,
		GenerateID:    generateUUID,
		AddToResponse: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate request ID
			requestID := r.Header.Get(cfg.HeaderName)
			if requestID == "" {
				requestID = cfg.GenerateID()
			}

			// Add to context
			ctx := context.WithValue(r.Context(), RequestIDContextKey, requestID)
			r = r.WithContext(ctx)

			// Add to response headers if enabled
			if cfg.AddToResponse {
				w.Header().Set(cfg.HeaderName, requestID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDOption is a functional option for RequestID middleware.
type RequestIDOption func(*RequestIDConfig)

// WithRequestIDHeader sets the header name for request IDs.
func WithRequestIDHeader(headerName string) RequestIDOption {
	return func(cfg *RequestIDConfig) {
		cfg.HeaderName = headerName
	}
}

// WithRequestIDGenerator sets a custom request ID generator.
func WithRequestIDGenerator(generator func() string) RequestIDOption {
	return func(cfg *RequestIDConfig) {
		cfg.GenerateID = generator
	}
}

// WithRequestIDResponse sets whether to add request ID to response headers.
func WithRequestIDResponse(addToResponse bool) RequestIDOption {
	return func(cfg *RequestIDConfig) {
		cfg.AddToResponse = addToResponse
	}
}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is found.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// generateUUID generates a UUID v4 string.
func generateUUID() string {
	return uuid.New().String()
}
