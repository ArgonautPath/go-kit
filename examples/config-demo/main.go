package main

import (
	"fmt"
	"os"

	"github.com/ArgonautPath/go-kit/pkg/config"
)

func main() {
	// Example 1: Basic configuration with defaults
	fmt.Println("=== Example 1: Basic Configuration ===")
	type AppConfig struct {
		Host  string `config:"env=APP_HOST,default=localhost"`
		Port  int    `config:"env=APP_PORT,default=8080"`
		Debug bool   `config:"env=APP_DEBUG,default=false"`
	}

	loader := config.NewLoader()
	var cfg AppConfig

	if err := loader.Load(&cfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Debug: %v\n\n", cfg.Debug)

	// Example 2: Configuration with environment variables
	fmt.Println("=== Example 2: Environment Variables ===")
	os.Setenv("APP_HOST", "example.com")
	os.Setenv("APP_PORT", "9090")
	defer func() {
		os.Unsetenv("APP_HOST")
		os.Unsetenv("APP_PORT")
	}()

	var cfg2 AppConfig
	if err := loader.Load(&cfg2); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s (from env)\n", cfg2.Host)
	fmt.Printf("Port: %d (from env)\n\n", cfg2.Port)

	// Example 3: Nested configuration
	fmt.Println("=== Example 3: Nested Configuration ===")
	type DatabaseConfig struct {
		Host     string `config:"env=DB_HOST,default=localhost"`
		Port     int    `config:"env=DB_PORT,default=5432"`
		Username string `config:"env=DB_USER,default=postgres"`
		Password string `config:"env=DB_PASS,required"`
	}

	type ServerConfig struct {
		Host string `config:"env=SERVER_HOST,default=0.0.0.0"`
		Port int    `config:"env=SERVER_PORT,default=8080,validate=range=1,65535"`
	}

	type FullAppConfig struct {
		Database DatabaseConfig
		Server   ServerConfig
		AppName  string `config:"env=APP_NAME,default=myapp"`
	}

	// Set required password
	os.Setenv("DB_PASS", "secret123")
	defer os.Unsetenv("DB_PASS")

	var fullCfg FullAppConfig
	if err := loader.Load(&fullCfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Database: %s:%d\n", fullCfg.Database.Host, fullCfg.Database.Port)
	fmt.Printf("Server: %s:%d\n", fullCfg.Server.Host, fullCfg.Server.Port)
	fmt.Printf("App Name: %s\n\n", fullCfg.AppName)

	// Example 4: Configuration with validation
	fmt.Println("=== Example 4: Configuration Validation ===")
	type ValidatedConfig struct {
		Email string `config:"env=APP_EMAIL,required,validate=email"`
		Port  int    `config:"env=APP_PORT,default=8080,validate=range=1,65535"`
		URL   string `config:"env=APP_URL,validate=url"`
	}

	// Test with valid configuration
	os.Setenv("APP_EMAIL", "user@example.com")
	os.Setenv("APP_URL", "https://example.com")
	defer func() {
		os.Unsetenv("APP_EMAIL")
		os.Unsetenv("APP_URL")
	}()

	var validatedCfg ValidatedConfig
	if err := loader.Load(&validatedCfg); err != nil {
		fmt.Printf("Validation Error: %v\n", err)
	} else {
		fmt.Printf("Email: %s (validated)\n", validatedCfg.Email)
		fmt.Printf("Port: %d (validated)\n", validatedCfg.Port)
		fmt.Printf("URL: %s (validated)\n\n", validatedCfg.URL)
	}

	// Example 5: Configuration with file loading
	fmt.Println("=== Example 5: File-based Configuration ===")
	// Create a temporary config file
	tmpFile := "/tmp/config-example.yaml"
	fileContent := `host: file-host
port: 3000
debug: true
`
	if err := os.WriteFile(tmpFile, []byte(fileContent), 0644); err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile)

	type FileConfig struct {
		Host  string `config:"env=APP_HOST,default=localhost"`
		Port  int    `config:"env=APP_PORT,default=8080"`
		Debug bool   `config:"env=APP_DEBUG,default=false"`
	}

	fileLoader := config.NewLoaderWithConfig(config.Config{
		FilePath:         tmpFile,
		ValidateAfterLoad: false,
	})

	var fileCfg FileConfig
	if err := fileLoader.Load(&fileCfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s (from file)\n", fileCfg.Host)
	fmt.Printf("Port: %d (from file)\n", fileCfg.Port)
	fmt.Printf("Debug: %v (from file)\n\n", fileCfg.Debug)

	// Example 6: Configuration priority (env > file > defaults)
	fmt.Println("=== Example 6: Configuration Priority ===")
	os.Setenv("APP_HOST", "env-host") // This should override file
	defer os.Unsetenv("APP_HOST")

	var priorityCfg FileConfig
	if err := fileLoader.Load(&priorityCfg); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Host: %s (env overrides file)\n", priorityCfg.Host)
	fmt.Printf("Port: %d (from file, env not set)\n", priorityCfg.Port)
}

