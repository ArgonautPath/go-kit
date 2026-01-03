# go-kit

A collection of reusable Go packages for building microservices and applications. This library provides essential utilities for HTTP clients, structured logging, and configuration management.

## Features

- **HTTP Client** (`pkg/httpclient`): Type-safe, generic HTTP client with functional options
- **Logger** (`pkg/logger`): Structured logging with async support, context correlation, and multiple output formats
- **Config** (`pkg/config`): Configuration loader with support for files, environment variables, and validation

## Installation

```bash
go get github.com/ArgonautPath/go-kit
```

## Packages

### HTTP Client

A type-safe HTTP client built with generics, supporting all HTTP methods with functional options.

```go
import "github.com/ArgonautPath/go-kit/pkg/httpclient"

// Create a client
client, err := httpclient.NewGeneric(httpclient.Config{
    BaseURL: "https://api.example.com",
    DefaultTimeout: 30 * time.Second,
})

// Make type-safe requests
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

var user User
resp, err := client.Get[User](ctx, "/users/1", 
    httpclient.WithHeaders(map[string]string{
        "Authorization": "Bearer token",
    }),
)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Body.Name) // Type-safe access
```

### Logger

Structured logging with support for JSON/text formats, async logging, context correlation, and caller information.

```go
import "github.com/ArgonautPath/go-kit/pkg/logger"

// Create a logger
log, err := logger.New(logger.Config{
    Level:  logger.InfoLevel,
    Output: logger.NewStdoutWriter(logger.JSONFormat),
    Format: logger.JSONFormat,
    AddCaller: true,
})

// Log with context
ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
log.Info(ctx, "request processed",
    logger.String("method", "GET"),
    logger.Int("status", 200),
    logger.Duration("duration", time.Millisecond*150),
)

// Create prefixed logger
httpLogger := log.Prefix("[HTTP]")
httpLogger.Info(ctx, "handling request")

// Async logging (non-blocking)
asyncLog, _ := logger.New(logger.Config{
    AsyncEnabled:  true,
    AsyncBufferSize: 1000,
    // ... other config
})
defer asyncLog.Close() // Flush on shutdown
```

### Config

Configuration loader with support for YAML/JSON files, environment variables, defaults, and validation.

```go
import "github.com/ArgonautPath/go-kit/pkg/config"

// Define configuration struct
type AppConfig struct {
    Host     string `config:"env=APP_HOST,default=localhost"`
    Port     int    `config:"env=APP_PORT,default=8080,validate=range=1,65535"`
    Database struct {
        Host     string `config:"env=DB_HOST,default=localhost"`
        Port     int    `config:"env=DB_PORT,default=5432"`
        Username string `config:"env=DB_USER,required"`
        Password string `config:"env=DB_PASS,required"`
    }
}

// Load configuration
loader := config.NewLoaderWithConfig(config.Config{
    FilePath:         "config.yaml",
    EnvPrefix:        "APP",
    ValidateAfterLoad: true,
})

var cfg AppConfig
if err := loader.Load(&cfg); err != nil {
    log.Fatal(err)
}
```

#### Configuration Tags

- `env=VAR_NAME`: Load from environment variable
- `default=value`: Default value if not set
- `required`: Field must be set
- `validate=email`: Validate email format
- `validate=url`: Validate URL format
- `validate=range=min,max`: Validate numeric range

## Examples

See the `examples/` directory for complete working examples:

- `examples/logger-demo/`: Logger usage example

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test ./... -cover
```

## Contributing

Contributions are welcome! Please ensure all tests pass and follow Go best practices.

## License

MIT License - see [LICENSE](LICENSE) file for details.

