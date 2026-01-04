//go:build !no_gin
// +build !no_gin

package middleware

import (
	"net/http"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/logger"
	"github.com/gin-gonic/gin"
)

// GinAdapter adapts a standard http.Handler middleware to work with Gin.
// This allows you to use go-kit middleware with Gin framework.
//
// Gin's middleware uses gin.HandlerFunc (func(*gin.Context)), while our
// middleware uses http.Handler. This adapter bridges the gap.
//
// Example:
//
//	import "github.com/ArgonautPath/go-kit/pkg/middleware"
//
//	r := gin.Default()
//	r.Use(middleware.GinAdapter(middleware.RequestID()))
//	r.Use(middleware.GinAdapter(middleware.Recovery()))
//	r.Use(middleware.GinAdapter(middleware.Logging(logger)))
func GinAdapter(m Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a handler that will execute the Gin context's next handlers
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Update the request in context
			c.Request = r
			// Continue with Gin's handler chain
			c.Next()
		})

		// Apply our middleware to the handler
		wrapped := m(handler)

		// Execute the wrapped handler with Gin's response writer and request
		wrapped.ServeHTTP(c.Writer, c.Request)
	}
}

// ToGinMiddleware converts a standard middleware to a Gin middleware.
// This is an alias for GinAdapter for better readability.
func ToGinMiddleware(m Middleware) gin.HandlerFunc {
	return GinAdapter(m)
}

// GinRequestID is a convenience function that returns a Gin middleware for request ID.
// It's equivalent to: GinAdapter(RequestID())
//
// Example:
//
//	r := gin.Default()
//	r.Use(middleware.GinRequestID())
//	r.Use(middleware.GinRequestID(middleware.WithRequestIDHeader("X-Custom-ID")))
func GinRequestID(opts ...RequestIDOption) gin.HandlerFunc {
	return GinAdapter(RequestID(opts...))
}

// GinRecovery is a convenience function that returns a Gin middleware for recovery.
// It's equivalent to: GinAdapter(Recovery())
//
// Example:
//
//	r := gin.Default()
//	r.Use(middleware.GinRecovery())
//	r.Use(middleware.GinRecovery(middleware.WithRecoveryPrintStack(true)))
func GinRecovery(opts ...RecoveryOption) gin.HandlerFunc {
	return GinAdapter(Recovery(opts...))
}

// GinLogging is a convenience function that returns a Gin middleware for logging.
// It requires a logger from the logger package.
//
// Example:
//
//	import (
//		"github.com/ArgonautPath/go-kit/pkg/logger"
//		"github.com/ArgonautPath/go-kit/pkg/middleware"
//	)
//
//	log, _ := logger.New(logger.Config{...})
//	r := gin.Default()
//	r.Use(middleware.GinLogging(log))
//	r.Use(middleware.GinLogging(log, middleware.WithSkipPaths("/health")))
func GinLogging(l logger.Logger, opts ...LoggingOption) gin.HandlerFunc {
	return GinAdapter(Logging(l, opts...))
}

// GinCORS is a convenience function that returns a Gin middleware for CORS.
// It's equivalent to: GinAdapter(CORS(cfg))
//
// Example:
//
//	r := gin.Default()
//	r.Use(middleware.GinCORS(middleware.CORSConfig{
//		AllowedOrigins: []string{"https://example.com"},
//		AllowedMethods: []string{"GET", "POST"},
//	}))
func GinCORS(cfg CORSConfig) gin.HandlerFunc {
	return GinAdapter(CORS(cfg))
}

// GinTimeout is a convenience function that returns a Gin middleware for timeout.
// It's equivalent to: GinAdapter(Timeout(timeout, opts...))
//
// Example:
//
//	r := gin.Default()
//	r.Use(middleware.GinTimeout(30 * time.Second))
//	r.Use(middleware.GinTimeout(30*time.Second, middleware.WithTimeoutMessage("Too slow")))
func GinTimeout(timeout time.Duration, opts ...TimeoutOption) gin.HandlerFunc {
	return GinAdapter(Timeout(timeout, opts...))
}
