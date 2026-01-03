package config_test

import (
	"fmt"
	"os"

	"github.com/ArgonautPath/go-kit/pkg/config"
)

// ExampleLoader_basic demonstrates basic configuration loading with defaults.
func ExampleLoader_basic() {
	// Define configuration struct
	type AppConfig struct {
		Host  string `config:"env=APP_HOST,default=localhost"`
		Port  int    `config:"env=APP_PORT,default=8080"`
		Debug bool   `config:"env=APP_DEBUG,default=false"`
	}

	// Create loader and load configuration
	loader := config.NewLoader()
	var cfg AppConfig

	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Debug: %v\n", cfg.Debug)

	// Output:
	// Host: localhost
	// Port: 8080
	// Debug: false
}

// ExampleLoader_withEnv demonstrates loading configuration from environment variables.
func ExampleLoader_withEnv() {
	// Set environment variables
	os.Setenv("APP_HOST", "example.com")
	os.Setenv("APP_PORT", "9090")
	defer func() {
		os.Unsetenv("APP_HOST")
		os.Unsetenv("APP_PORT")
	}()

	type AppConfig struct {
		Host string `config:"env=APP_HOST,default=localhost"`
		Port int    `config:"env=APP_PORT,default=8080"`
	}

	loader := config.NewLoader()
	var cfg AppConfig

	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)

	// Output:
	// Host: example.com
	// Port: 9090
}

// ExampleLoader_nested demonstrates loading nested configuration structures.
func ExampleLoader_nested() {
	type DatabaseConfig struct {
		Host     string `config:"env=DB_HOST,default=localhost"`
		Port     int    `config:"env=DB_PORT,default=5432"`
		Username string `config:"env=DB_USER,default=postgres"`
		Password string `config:"env=DB_PASS,required"`
	}

	type ServerConfig struct {
		Host string `config:"env=SERVER_HOST,default=0.0.0.0"`
		Port int    `config:"env=SERVER_PORT,default=8080"`
	}

	type AppConfig struct {
		Database DatabaseConfig
		Server   ServerConfig
	}

	// Set required environment variable
	os.Setenv("DB_PASS", "secret123")
	defer os.Unsetenv("DB_PASS")

	loader := config.NewLoader()
	var cfg AppConfig

	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Database: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)

	// Output:
	// Database: localhost:5432
	// Server: 0.0.0.0:8080
}

// ExampleLoader_fromFile demonstrates loading configuration from a YAML file.
func ExampleLoader_fromFile() {
	// Create a temporary YAML file
	tmpFile := "/tmp/config_example.yaml"
	content := `
host: file-host
port: 3000
database:
  host: db.example.com
  port: 3306
`
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	type DatabaseConfig struct {
		Host string
		Port int
	}

	type AppConfig struct {
		Host     string
		Port     int
		Database DatabaseConfig
	}

	loader := config.NewLoader()
	var cfg AppConfig

	// Load from specific file
	if err := loader.LoadFromFile(tmpFile, &cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Database: %s:%d\n", cfg.Database.Host, cfg.Database.Port)

	// Output:
	// Host: file-host
	// Port: 3000
	// Database: db.example.com:3306
}

// ExampleLoader_validation demonstrates configuration validation.
func ExampleLoader_validation() {
	type ValidatedConfig struct {
		Email string `config:"env=EMAIL,required,validate=email"`
		Port  int    `config:"env=PORT,default=8080,validate=range=1,65535"`
		URL   string `config:"env=URL,validate=url"`
	}

	// Set valid environment variables
	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("PORT", "9090")
	os.Setenv("URL", "https://example.com")
	defer func() {
		os.Unsetenv("EMAIL")
		os.Unsetenv("PORT")
		os.Unsetenv("URL")
	}()

	loader := config.NewLoader()
	var cfg ValidatedConfig

	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	fmt.Printf("Email: %s\n", cfg.Email)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("URL: %s\n", cfg.URL)

	// Output:
	// Email: user@example.com
	// Port: 9090
	// URL: https://example.com
}

// ExampleLoader_priority demonstrates source priority (env overrides file).
func ExampleLoader_priority() {
	// Create a temporary config file
	tmpFile := "/tmp/config_priority.yaml"
	content := `
host: file-host
port: 2000
`
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	// Set environment variable (should override file)
	os.Setenv("APP_HOST", "env-host")
	defer os.Unsetenv("APP_HOST")

	type AppConfig struct {
		Host string `config:"env=APP_HOST,default=localhost"`
		Port int    `config:"env=APP_PORT,default=8080"`
	}

	loader := config.NewLoaderWithConfig(config.Config{
		FilePath:          tmpFile,
		ValidateAfterLoad: false,
	})

	var cfg AppConfig
	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Host comes from env (highest priority)
	// Port comes from file (env not set)
	fmt.Printf("Host: %s (from env)\n", cfg.Host)
	fmt.Printf("Port: %d (from file)\n", cfg.Port)

	// Output:
	// Host: env-host (from env)
	// Port: 2000 (from file)
}

// ExampleLoader_withPrefix demonstrates using environment variable prefix.
func ExampleLoader_withPrefix() {
	// Set prefixed environment variables
	os.Setenv("MYAPP_HOST", "prefixed-host")
	os.Setenv("MYAPP_PORT", "7777")
	defer func() {
		os.Unsetenv("MYAPP_HOST")
		os.Unsetenv("MYAPP_PORT")
	}()

	type AppConfig struct {
		Host string `config:"env=HOST"`
		Port int    `config:"env=PORT"`
	}

	loader := config.NewLoaderWithConfig(config.Config{
		EnvPrefix:         "MYAPP",
		ValidateAfterLoad: false,
	})

	var cfg AppConfig
	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)

	// Output:
	// Host: prefixed-host
	// Port: 7777
}
