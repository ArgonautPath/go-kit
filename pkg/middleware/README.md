# Middleware Package

HTTP middleware utilities for Go applications. This package provides common middleware patterns for building robust HTTP servers.

## Features

- **RequestID**: Injects unique request IDs for request tracing
- **Logging**: Structured logging of HTTP requests/responses
- **Recovery**: Panic recovery to prevent server crashes
- **CORS**: Cross-Origin Resource Sharing support
- **Timeout**: Request timeout enforcement
- **Chain**: Compose multiple middlewares together

## Installation

```bash
go get github.com/ArgonautPath/go-kit/pkg/middleware
```

## Usage with Standard net/http

```go
import (
    "net/http"
    "github.com/ArgonautPath/go-kit/pkg/middleware"
    "github.com/ArgonautPath/go-kit/pkg/logger"
)

// Create logger
log, _ := logger.New(logger.Config{...})

// Create middleware chain
chain := middleware.Chain(
    middleware.RequestID(),
    middleware.Recovery(),
    middleware.Logging(log),
    middleware.CORS(middleware.CORSConfig{
        AllowedOrigins: []string{"https://example.com"},
    }),
    middleware.Timeout(30 * time.Second),
)

// Apply to handler
handler := chain(yourHandler)
http.ListenAndServe(":8080", handler)
```

## Usage with Gin Framework

The middleware package includes Gin adapters to use with the [Gin](https://github.com/gin-gonic/gin) framework.

**Note**: To use Gin adapters, you need to add Gin as a dependency:

```bash
go get github.com/gin-gonic/gin
```

### Using GinAdapter

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/ArgonautPath/go-kit/pkg/middleware"
    "github.com/ArgonautPath/go-kit/pkg/logger"
)

r := gin.Default()

// Use GinAdapter to convert standard middleware to Gin middleware
r.Use(middleware.GinAdapter(middleware.RequestID()))
r.Use(middleware.GinAdapter(middleware.Recovery()))
r.Use(middleware.GinAdapter(middleware.Logging(log)))
```

### Using Convenience Functions

```go
r := gin.Default()

// Convenience functions for each middleware
r.Use(middleware.GinRequestID())
r.Use(middleware.GinRecovery())
r.Use(middleware.GinLogging(log))
r.Use(middleware.GinCORS(middleware.CORSConfig{
    AllowedOrigins: []string{"https://example.com"},
}))
r.Use(middleware.GinTimeout(30 * time.Second))
```

### Complete Gin Example

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/ArgonautPath/go-kit/pkg/logger"
    "github.com/ArgonautPath/go-kit/pkg/middleware"
)

func main() {
    log, _ := logger.New(logger.Config{...})
    
    r := gin.Default()
    
    // Apply middlewares
    r.Use(middleware.GinRequestID())
    r.Use(middleware.GinRecovery())
    r.Use(middleware.GinLogging(log))
    r.Use(middleware.GinCORS(middleware.CORSConfig{
        AllowedOrigins: []string{"*"},
    }))
    
    // Routes
    r.GET("/", func(c *gin.Context) {
        requestID := middleware.GetRequestID(c.Request.Context())
        c.JSON(http.StatusOK, gin.H{
            "request_id": requestID,
            "message": "Hello, World!",
        })
    })
    
    r.Run(":8080")
}
```

## Building Without Gin

If you don't use Gin, you can build the package without Gin support using the build tag:

```bash
go build -tags=no_gin ./pkg/middleware/...
```

This excludes the Gin adapter code from compilation.

## Middleware Details

### RequestID

Injects a unique request ID into each request for tracing.

```go
middleware.RequestID(
    middleware.WithRequestIDHeader("X-Request-ID"),
    middleware.WithRequestIDResponse(true),
)
```

### Logging

Logs HTTP requests and responses with structured logging.

```go
middleware.Logging(log,
    middleware.WithSkipPaths("/health", "/metrics"),
    middleware.WithSkipStatusCodes(200),
)
```

### Recovery

Recovers from panics and prevents server crashes.

```go
middleware.Recovery(
    middleware.WithRecoveryPrintStack(false),
    middleware.WithRecoveryHandler(customHandler),
)
```

### CORS

Handles Cross-Origin Resource Sharing headers.

```go
middleware.CORS(middleware.CORSConfig{
    AllowedOrigins: []string{"https://example.com"},
    AllowedMethods: []string{"GET", "POST"},
    AllowedHeaders: []string{"Content-Type"},
    AllowCredentials: true,
})
```

### Timeout

Enforces request timeouts.

```go
middleware.Timeout(30 * time.Second,
    middleware.WithTimeoutMessage("Request timeout"),
    middleware.WithTimeoutStatusCode(http.StatusRequestTimeout),
)
```

## See Also

- [Examples](../examples/middleware-demo/) - Complete usage examples
- [Logger Package](../logger/) - Structured logging
- [HTTP Client Package](../httpclient/) - HTTP client utilities

