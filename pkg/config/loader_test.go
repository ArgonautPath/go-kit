package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig represents a test configuration struct.
type TestConfig struct {
	Host     string `config:"env=TEST_HOST,default=localhost"`
	Port     int    `config:"env=TEST_PORT,default=8080"`
	Database struct {
		Host     string `config:"env=DB_HOST,default=localhost"`
		Port     int    `config:"env=DB_PORT,default=5432"`
		Username string `config:"env=DB_USER,required"`
		Password string `config:"env=DB_PASS,required"`
	}
}

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Error("NewLoader() returned nil")
	}
}

func TestLoader_SetDefaults(t *testing.T) {
	loader := NewLoader()

	var cfg TestConfig
	if err := loader.SetDefaults(&cfg); err != nil {
		t.Fatalf("SetDefaults() error = %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("cfg.Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("cfg.Database.Host = %q, want %q", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("cfg.Database.Port = %d, want %d", cfg.Database.Port, 5432)
	}
}

func TestLoader_LoadFromEnv(t *testing.T) {
	loader := NewLoader()

	// Set environment variables
	os.Setenv("TEST_HOST", "example.com")
	os.Setenv("TEST_PORT", "9090")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASS", "testpass")
	defer func() {
		os.Unsetenv("TEST_HOST")
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASS")
	}()

	var cfg TestConfig
	if err := loader.LoadFromEnv(&cfg); err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "example.com")
	}
	if cfg.Port != 9090 {
		t.Errorf("cfg.Port = %d, want %d", cfg.Port, 9090)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("cfg.Database.Host = %q, want %q", cfg.Database.Host, "db.example.com")
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("cfg.Database.Port = %d, want %d", cfg.Database.Port, 3306)
	}
	if cfg.Database.Username != "testuser" {
		t.Errorf("cfg.Database.Username = %q, want %q", cfg.Database.Username, "testuser")
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("cfg.Database.Password = %q, want %q", cfg.Database.Password, "testpass")
	}
}

func TestLoader_LoadFromFile_YAML(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")
	content := `
host: yaml-host
port: 3000
database:
  host: yaml-db-host
  port: 5433
  username: yaml-user
  password: yaml-pass
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()

	var cfg TestConfig
	if err := loader.LoadFromFile(filePath, &cfg); err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.Host != "yaml-host" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "yaml-host")
	}
	if cfg.Port != 3000 {
		t.Errorf("cfg.Port = %d, want %d", cfg.Port, 3000)
	}
	if cfg.Database.Host != "yaml-db-host" {
		t.Errorf("cfg.Database.Host = %q, want %q", cfg.Database.Host, "yaml-db-host")
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("cfg.Database.Port = %d, want %d", cfg.Database.Port, 5433)
	}
}

func TestLoader_LoadFromFile_JSON(t *testing.T) {
	// Create temporary JSON file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.json")
	content := `{
  "host": "json-host",
  "port": 4000,
  "database": {
    "host": "json-db-host",
    "port": 5434,
    "username": "json-user",
    "password": "json-pass"
  }
}`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()

	var cfg TestConfig
	if err := loader.LoadFromFile(filePath, &cfg); err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if cfg.Host != "json-host" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "json-host")
	}
	if cfg.Port != 4000 {
		t.Errorf("cfg.Port = %d, want %d", cfg.Port, 4000)
	}
	if cfg.Database.Host != "json-db-host" {
		t.Errorf("cfg.Database.Host = %q, want %q", cfg.Database.Host, "json-db-host")
	}
	if cfg.Database.Port != 5434 {
		t.Errorf("cfg.Database.Port = %d, want %d", cfg.Database.Port, 5434)
	}
}

func TestLoader_Load_Priority(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")
	content := `
host: file-host
port: 2000
database:
  host: file-db-host
  port: 5435
  username: file-user
  password: file-pass
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set environment variables (should override file)
	os.Setenv("TEST_HOST", "env-host")
	os.Setenv("DB_USER", "env-user")
	defer func() {
		os.Unsetenv("TEST_HOST")
		os.Unsetenv("DB_USER")
	}()

	loader := NewLoaderWithConfig(Config{
		FilePath:          filePath,
		ValidateAfterLoad: false, // Disable validation for this test
	})

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Env should override file
	if cfg.Host != "env-host" {
		t.Errorf("cfg.Host = %q, want %q (env should override file)", cfg.Host, "env-host")
	}

	// File value should be used if env not set
	if cfg.Port != 2000 {
		t.Errorf("cfg.Port = %d, want %d (from file)", cfg.Port, 2000)
	}

	// Env should override nested struct
	if cfg.Database.Username != "env-user" {
		t.Errorf("cfg.Database.Username = %q, want %q (env should override file)", cfg.Database.Username, "env-user")
	}

	// File value should be used if env not set
	if cfg.Database.Host != "file-db-host" {
		t.Errorf("cfg.Database.Host = %q, want %q (from file)", cfg.Database.Host, "file-db-host")
	}
}

func TestLoader_Load_Defaults(t *testing.T) {
	loader := NewLoaderWithConfig(Config{
		FilePath:          "", // No file
		ValidateAfterLoad: false,
	})

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have default values
	if cfg.Host != "localhost" {
		t.Errorf("cfg.Host = %q, want %q (default)", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("cfg.Port = %d, want %d (default)", cfg.Port, 8080)
	}
}

func TestValidateStruct_Required(t *testing.T) {
	type RequiredConfig struct {
		Name     string `config:"required"`
		Optional string
	}

	tests := []struct {
		name    string
		cfg     RequiredConfig
		wantErr bool
	}{
		{
			name:    "missing required field",
			cfg:     RequiredConfig{},
			wantErr: true,
		},
		{
			name:    "required field set",
			cfg:     RequiredConfig{Name: "test"},
			wantErr: false,
		},
		{
			name:    "empty string required field",
			cfg:     RequiredConfig{Name: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct_Email(t *testing.T) {
	type EmailConfig struct {
		Email string `config:"validate=email"`
	}

	tests := []struct {
		name    string
		cfg     EmailConfig
		wantErr bool
	}{
		{
			name:    "valid email",
			cfg:     EmailConfig{Email: "test@example.com"},
			wantErr: false,
		},
		{
			name:    "invalid email",
			cfg:     EmailConfig{Email: "invalid-email"},
			wantErr: true,
		},
		{
			name:    "empty email (not required)",
			cfg:     EmailConfig{Email: ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct_URL(t *testing.T) {
	type URLConfig struct {
		URL string `config:"validate=url"`
	}

	tests := []struct {
		name    string
		cfg     URLConfig
		wantErr bool
	}{
		{
			name:    "valid URL",
			cfg:     URLConfig{URL: "https://example.com"},
			wantErr: false,
		},
		{
			name:    "invalid URL",
			cfg:     URLConfig{URL: "not-a-url"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct_Range(t *testing.T) {
	type RangeConfig struct {
		Port int `config:"validate=range=1,65535"`
	}

	tests := []struct {
		name    string
		cfg     RangeConfig
		wantErr bool
	}{
		{
			name:    "valid port",
			cfg:     RangeConfig{Port: 8080},
			wantErr: false,
		},
		{
			name:    "port too low",
			cfg:     RangeConfig{Port: 0},
			wantErr: true,
		},
		{
			name:    "port too high",
			cfg:     RangeConfig{Port: 70000},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoader_Load_Validation(t *testing.T) {
	type ValidatedConfig struct {
		Email string `config:"env=TEST_EMAIL,required,validate=email"`
		Port  int    `config:"env=TEST_PORT,validate=range=1,65535"`
	}

	loader := NewLoaderWithConfig(Config{
		ValidateAfterLoad: true,
	})

	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		wantErr bool
	}{
		{
			name: "valid configuration",
			setup: func() {
				os.Setenv("TEST_EMAIL", "test@example.com")
				os.Setenv("TEST_PORT", "8080")
			},
			cleanup: func() {
				os.Unsetenv("TEST_EMAIL")
				os.Unsetenv("TEST_PORT")
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			setup: func() {
				os.Unsetenv("TEST_EMAIL")
				os.Setenv("TEST_PORT", "8080")
			},
			cleanup: func() {
				os.Unsetenv("TEST_PORT")
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			setup: func() {
				os.Setenv("TEST_EMAIL", "invalid-email")
				os.Setenv("TEST_PORT", "8080")
			},
			cleanup: func() {
				os.Unsetenv("TEST_EMAIL")
				os.Unsetenv("TEST_PORT")
			},
			wantErr: true,
		},
		{
			name: "port out of range",
			setup: func() {
				os.Setenv("TEST_EMAIL", "test@example.com")
				os.Setenv("TEST_PORT", "70000")
			},
			cleanup: func() {
				os.Unsetenv("TEST_EMAIL")
				os.Unsetenv("TEST_PORT")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			var cfg ValidatedConfig
			err := loader.Load(&cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     Format
	}{
		{"yaml extension", "config.yaml", YAMLFormat},
		{"yml extension", "config.yml", YAMLFormat},
		{"json extension", "config.json", JSONFormat},
		{"unknown extension", "config.txt", UnknownFormat},
		{"no extension", "config", UnknownFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectFormat(tt.filename); got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.txt")
	if err := os.WriteFile(filePath, []byte("invalid"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var cfg TestConfig
	err := DecodeFile(filePath, &cfg)
	if err == nil {
		t.Error("DecodeFile() should return error for unknown format")
	}
}

func TestDecodeFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")
	content := "invalid: yaml: content: ["
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var cfg TestConfig
	err := DecodeFile(filePath, &cfg)
	if err == nil {
		t.Error("DecodeFile() should return error for invalid YAML")
	}
}

func TestDecodeFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.json")
	content := `{"invalid": json}`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var cfg TestConfig
	err := DecodeFile(filePath, &cfg)
	if err == nil {
		t.Error("DecodeFile() should return error for invalid JSON")
	}
}

func TestLoader_LoadFromFile_NonExistent(t *testing.T) {
	loader := NewLoader()

	var cfg TestConfig
	err := loader.LoadFromFile("/nonexistent/file.yaml", &cfg)
	if err == nil {
		t.Error("LoadFromFile() should return error for non-existent file")
	}
}

func TestEnvSource_Prefix(t *testing.T) {
	type PrefixedConfig struct {
		Host string `config:"env=HOST"`
		Port int    `config:"env=PORT"`
	}

	os.Setenv("APP_HOST", "prefixed-host")
	os.Setenv("APP_PORT", "9999")
	defer func() {
		os.Unsetenv("APP_HOST")
		os.Unsetenv("APP_PORT")
	}()

	source := NewEnvSource("APP")
	var cfg PrefixedConfig
	if err := source.Load(&cfg); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Host != "prefixed-host" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "prefixed-host")
	}
	if cfg.Port != 9999 {
		t.Errorf("cfg.Port = %d, want %d", cfg.Port, 9999)
	}
}

func TestDefaultSource_NestedStruct(t *testing.T) {
	type NestedConfig struct {
		Outer struct {
			Inner struct {
				Value string `config:"default=nested-default"`
			}
		}
	}

	source := NewDefaultSource()
	var cfg NestedConfig
	if err := source.Load(&cfg); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Outer.Inner.Value != "nested-default" {
		t.Errorf("cfg.Outer.Inner.Value = %q, want %q", cfg.Outer.Inner.Value, "nested-default")
	}
}
