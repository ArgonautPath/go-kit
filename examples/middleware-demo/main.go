package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/logger"
	"github.com/ArgonautPath/go-kit/pkg/middleware"
)

func main() {
	// Create a logger
	log, _ := logger.New(logger.Config{
		Level:     logger.InfoLevel,
		Output:    logger.NewStdoutWriter(logger.JSONFormat),
		Format:    logger.JSONFormat,
		AddCaller: true,
	})

	// Create a simple HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from context
		requestID := middleware.GetRequestID(r.Context())

		log.Info(r.Context(), "Handling request",
			logger.String("request_id", requestID),
			logger.String("path", r.URL.Path),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Hello, World!", "request_id": "%s"}`, requestID)
	})

	// Example 1: Basic middleware chain
	fmt.Println("=== Example 1: Basic Middleware Chain ===")
	chain1 := middleware.Chain(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.Logging(log),
	)
	handler1 := chain1(handler)

	// Example 2: CORS middleware
	fmt.Println("\n=== Example 2: CORS Middleware ===")
	corsHandler := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins:   []string{"https://example.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	})(handler1)

	// Example 3: Timeout middleware
	fmt.Println("\n=== Example 3: Timeout Middleware ===")
	_ = middleware.Timeout(30*time.Second,
		middleware.WithTimeoutMessage("Request took too long"),
		middleware.WithTimeoutStatusCode(http.StatusRequestTimeout),
	)(corsHandler)

	// Example 4: Complete middleware stack
	fmt.Println("\n=== Example 4: Complete Middleware Stack ===")
	completeStack := middleware.Chain(
		middleware.RequestID(
			middleware.WithRequestIDHeader("X-Request-ID"),
			middleware.WithRequestIDResponse(true),
		),
		middleware.Recovery(
			middleware.WithRecoveryPrintStack(false),
		),
		middleware.Logging(log,
			middleware.WithSkipPaths("/health", "/metrics"),
			middleware.WithSkipStatusCodes(200),
		),
		middleware.CORS(middleware.CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		}),
		middleware.Timeout(30*time.Second),
	)

	finalHandler := completeStack(handler)

	// Example 5: Panic recovery demonstration
	fmt.Println("\n=== Example 5: Panic Recovery ===")
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("This panic will be recovered by the Recovery middleware")
	})

	recoveredHandler := middleware.Recovery(
		middleware.WithRecoveryHandler(func(w http.ResponseWriter, r *http.Request, err interface{}) {
			log.Error(r.Context(), "Panic recovered", fmt.Errorf("%v", err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "Internal server error"}`)
		}),
	)(panicHandler)

	// Start server with complete middleware stack
	fmt.Println("\n=== Starting HTTP Server ===")
	fmt.Println("Server listening on :8080")
	fmt.Println("Try: curl http://localhost:8080/")
	fmt.Println("Try: curl -H 'Origin: http://localhost:3000' http://localhost:8080/")

	mux := http.NewServeMux()
	mux.Handle("/", finalHandler)
	mux.Handle("/panic", recoveredHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok"}`)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// In a real application, you would use server.ListenAndServe()
	// For demo purposes, we'll just show the setup
	fmt.Println("\nServer configured with middleware stack:")
	fmt.Println("  1. RequestID - Adds unique request ID to each request")
	fmt.Println("  2. Recovery - Recovers from panics")
	fmt.Println("  3. Logging - Logs all HTTP requests")
	fmt.Println("  4. CORS - Handles cross-origin requests")
	fmt.Println("  5. Timeout - Enforces request timeout")

	// Uncomment to actually start the server:
	// if err := server.ListenAndServe(); err != nil {
	// 	log.Fatal(err)
	// }

	_ = server // Suppress unused variable warning
}
