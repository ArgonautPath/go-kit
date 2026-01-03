package config

import (
	"fmt"
)

// Loader defines the interface for loading configuration.
type Loader interface {
	Load(cfg interface{}) error
	LoadFromFile(path string, cfg interface{}) error
	LoadFromEnv(cfg interface{}) error
	SetDefaults(cfg interface{}) error
}

// Config holds configuration for the loader.
type Config struct {
	// FilePath is the path to the configuration file (optional).
	FilePath string
	// EnvPrefix is the prefix for environment variables (optional).
	EnvPrefix string
	// ValidateAfterLoad enables validation after loading (default: true).
	ValidateAfterLoad bool
}

// loader is the concrete implementation of Loader.
type loader struct {
	config Config
}

// NewLoader creates a new loader with default configuration.
func NewLoader() Loader {
	return NewLoaderWithConfig(Config{
		ValidateAfterLoad: true,
	})
}

// NewLoaderWithConfig creates a new loader with the given configuration.
func NewLoaderWithConfig(cfg Config) Loader {
	return &loader{
		config: cfg,
	}
}

// Load loads configuration from multiple sources with priority:
// 1. File (if FilePath is set and file exists)
// 2. Environment variables
// 3. Default values from struct tags
func (l *loader) Load(cfg interface{}) error {
	// Step 1: Apply defaults first (lowest priority)
	if err := l.SetDefaults(cfg); err != nil {
		return fmt.Errorf("set defaults: %w", err)
	}

	// Step 2: Load from file (if specified)
	if l.config.FilePath != "" {
		if err := l.LoadFromFile(l.config.FilePath, cfg); err != nil {
			return fmt.Errorf("load from file: %w", err)
		}
	}

	// Step 3: Load from environment variables (highest priority, overrides file)
	if err := l.LoadFromEnv(cfg); err != nil {
		return fmt.Errorf("load from env: %w", err)
	}

	// Step 4: Validate if enabled
	if l.config.ValidateAfterLoad {
		if err := ValidateStruct(cfg); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// LoadFromFile loads configuration from a specific file.
func (l *loader) LoadFromFile(path string, cfg interface{}) error {
	source := NewFileSource(path)
	return source.Load(cfg)
}

// LoadFromEnv loads configuration from environment variables.
func (l *loader) LoadFromEnv(cfg interface{}) error {
	source := NewEnvSource(l.config.EnvPrefix)
	return source.Load(cfg)
}

// SetDefaults applies default values from struct tags.
func (l *loader) SetDefaults(cfg interface{}) error {
	source := NewDefaultSource()
	return source.Load(cfg)
}

// fileNotFoundError represents a file not found error.
type fileNotFoundError struct {
	Path string
}

// Error implements the error interface.
func (e *fileNotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.Path)
}
