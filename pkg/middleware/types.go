package middleware

import (
	"net/http"
)

// Middleware is a function that wraps an HTTP handler.
// It receives the next handler in the chain and returns a new handler.
type Middleware func(http.Handler) http.Handler

// HandlerFunc is a function type that matches http.HandlerFunc.
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Chain chains multiple middlewares together.
// Middlewares are executed in the order they are provided.
// The first middleware in the slice is the outermost (executed first),
// and the last middleware is the innermost (executed last).
//
// Example:
//
//	chain := Chain(
//		RequestID(),
//		Recovery(),
//		Logging(logger),
//	)
//	handler := chain(finalHandler)
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		// Apply middlewares in reverse order so the first one in the slice
		// is the outermost (executed first)
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Apply applies a middleware to an http.Handler.
func Apply(m Middleware, handler http.Handler) http.Handler {
	return m(handler)
}

// ApplyFunc applies a middleware to an http.HandlerFunc.
func ApplyFunc(m Middleware, handler HandlerFunc) http.Handler {
	return m(http.HandlerFunc(handler))
}
