package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins. Use "*" to allow all origins.
	AllowedOrigins []string
	// AllowedMethods is a list of allowed HTTP methods.
	AllowedMethods []string
	// AllowedHeaders is a list of allowed headers.
	AllowedHeaders []string
	// ExposedHeaders is a list of headers that can be exposed to the client.
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials can be included in requests.
	AllowCredentials bool
	// MaxAge is the maximum age for preflight requests in seconds.
	MaxAge int
}

// CORS handles Cross-Origin Resource Sharing (CORS) headers.
// It supports preflight OPTIONS requests and adds appropriate CORS headers
// to all responses.
//
// Example:
//
//	mux := http.NewServeMux()
//	handler := CORS(CORSConfig{
//		AllowedOrigins: []string{"https://example.com"},
//		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
//		AllowedHeaders: []string{"Content-Type", "Authorization"},
//	})(mux)
func CORS(cfg CORSConfig) Middleware {
	// Set defaults
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 86400 // 24 hours
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Handle preflight request
			if r.Method == http.MethodOptions {
				// Check if origin is allowed
				if isOriginAllowed(origin, cfg.AllowedOrigins) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
					w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))

					if cfg.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}

					if len(cfg.ExposedHeaders) > 0 {
						w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
					}

					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			// Handle actual request
			if isOriginAllowed(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if len(cfg.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if an origin is allowed.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	// Allow all origins
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return true
	}

	// Check if origin is in allowed list
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	return false
}

