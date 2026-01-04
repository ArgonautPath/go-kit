package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// RecoveryConfig holds configuration for the Recovery middleware.
type RecoveryConfig struct {
	// Handler is called when a panic occurs. If nil, a default handler is used.
	Handler func(http.ResponseWriter, *http.Request, interface{})
	// PrintStack prints the stack trace to the response.
	PrintStack bool
	// StackSize limits the size of the printed stack trace.
	StackSize int
}

// Recovery recovers from panics and returns a 500 Internal Server Error.
// It prevents the server from crashing and optionally logs the panic.
//
// Example:
//
//	mux := http.NewServeMux()
//	handler := Recovery()(mux)
func Recovery(opts ...RecoveryOption) Middleware {
	cfg := RecoveryConfig{
		Handler:    defaultRecoveryHandler,
		PrintStack: false,
		StackSize:  1024 * 1024, // 1MB
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Call custom handler if provided
					if cfg.Handler != nil {
						cfg.Handler(w, r, err)
					} else {
						// Default handler
						w.WriteHeader(http.StatusInternalServerError)
						w.Header().Set("Content-Type", "text/plain; charset=utf-8")
						fmt.Fprintf(w, "Internal Server Error\n")

						if cfg.PrintStack {
							stack := debug.Stack()
							if len(stack) > cfg.StackSize {
								stack = stack[:cfg.StackSize]
							}
							fmt.Fprintf(w, "\nStack Trace:\n%s", stack)
						}
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryOption is a functional option for Recovery middleware.
type RecoveryOption func(*RecoveryConfig)

// WithRecoveryHandler sets a custom panic handler.
func WithRecoveryHandler(handler func(http.ResponseWriter, *http.Request, interface{})) RecoveryOption {
	return func(cfg *RecoveryConfig) {
		cfg.Handler = handler
	}
}

// WithRecoveryPrintStack enables printing the stack trace in the response.
func WithRecoveryPrintStack(enabled bool) RecoveryOption {
	return func(cfg *RecoveryConfig) {
		cfg.PrintStack = enabled
	}
}

// WithRecoveryStackSize sets the maximum stack trace size to print.
func WithRecoveryStackSize(size int) RecoveryOption {
	return func(cfg *RecoveryConfig) {
		cfg.StackSize = size
	}
}

// defaultRecoveryHandler is the default panic handler.
func defaultRecoveryHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Internal Server Error\n")
}
