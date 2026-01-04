package main

import (
	"net/http"
	"time"

	"github.com/ArgonautPath/go-kit/pkg/logger"
	"github.com/ArgonautPath/go-kit/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a logger
	log, _ := logger.New(logger.Config{
		Level:     logger.InfoLevel,
		Output:    logger.NewStdoutWriter(logger.JSONFormat),
		Format:    logger.JSONFormat,
		AddCaller: true,
	})

	// Create Gin router
	r := gin.Default()

	// Example 1: Using GinAdapter to use standard middleware
	r.Use(middleware.GinAdapter(middleware.RequestID()))
	r.Use(middleware.GinAdapter(middleware.Recovery()))
	r.Use(middleware.GinAdapter(middleware.Logging(log)))

	// Example 2: Using convenience functions
	r.Use(middleware.GinRequestID())
	r.Use(middleware.GinRecovery())
	r.Use(middleware.GinLogging(log))

	// Example 3: Using CORS middleware
	r.Use(middleware.GinCORS(middleware.CORSConfig{
		AllowedOrigins:   []string{"https://example.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	// Example 4: Using Timeout middleware
	r.Use(middleware.GinTimeout(30*time.Second,
		middleware.WithTimeoutMessage("Request timeout"),
	))

	// Example routes
	r.GET("/", func(c *gin.Context) {
		// Get request ID from context
		requestID := middleware.GetRequestID(c.Request.Context())

		c.JSON(http.StatusOK, gin.H{
			"message":    "Hello, World!",
			"request_id": requestID,
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("This will be recovered by Recovery middleware")
	})

	// Start server
	r.Run(":8080")
}
