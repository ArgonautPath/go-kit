package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Source defines the interface for configuration sources.
type Source interface {
	Load(cfg interface{}) error
}

// FileSource loads configuration from a file.
type FileSource struct {
	Path string
}

// NewFileSource creates a new file source.
func NewFileSource(path string) *FileSource {
	return &FileSource{Path: path}
}

// Load loads configuration from a file.
func (s *FileSource) Load(cfg interface{}) error {
	if s.Path == "" {
		return nil // No file specified, skip
	}

	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", s.Path)
	}

	return DecodeFile(s.Path, cfg)
}

// EnvSource loads configuration from environment variables.
type EnvSource struct {
	Prefix string
}

// NewEnvSource creates a new environment variable source.
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{Prefix: prefix}
}

// Load loads configuration from environment variables.
func (s *EnvSource) Load(cfg interface{}) error {
	return s.loadStruct(cfg, "")
}

// loadStruct recursively loads environment variables into a struct.
func (s *EnvSource) loadStruct(cfg interface{}, prefix string) error {
	rv := reflect.ValueOf(cfg)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct or pointer to struct")
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get config tag
		tag := field.Tag.Get("config")
		options := parseTagOptions(tag)

		// Get environment variable name
		envKey := s.getEnvKey(field, prefix, options)

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := s.loadStruct(fieldValue.Addr().Interface(), envKey); err != nil {
				return err
			}
			continue
		}

		// Skip if no env tag and no prefix-based mapping
		if options["env"] == "" && envKey == "" {
			continue
		}

		// Get value from environment
		envValue := os.Getenv(envKey)
		if envValue == "" {
			continue // No env var set, skip
		}

		// Set the field value
		if err := s.setFieldValue(fieldValue, envValue); err != nil {
			return fmt.Errorf("field %q: %w", field.Name, err)
		}
	}

	return nil
}

// getEnvKey determines the environment variable key for a field.
func (s *EnvSource) getEnvKey(field reflect.StructField, prefix string, options map[string]string) string {
	// Check for explicit env tag
	if envKey := options["env"]; envKey != "" {
		// If prefix is set, prepend it to the env key
		if s.Prefix != "" {
			return fmt.Sprintf("%s_%s", s.Prefix, strings.ToUpper(envKey))
		}
		return envKey
	}

	// Build key from prefix and field name
	fieldName := field.Name
	key := camelToSnake(fieldName)
	upperKey := strings.ToUpper(key)
	
	if s.Prefix != "" {
		return fmt.Sprintf("%s_%s", s.Prefix, upperKey)
	}

	return upperKey
}

// camelToSnake converts camelCase to snake_case.
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// setFieldValue sets a field value from a string.
func (s *EnvSource) setFieldValue(fieldValue reflect.Value, value string) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer: %w", err)
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %w", err)
		}
		fieldValue.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %w", err)
		}
		fieldValue.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	return nil
}

// DefaultSource applies default values from struct tags.
type DefaultSource struct{}

// NewDefaultSource creates a new default value source.
func NewDefaultSource() *DefaultSource {
	return &DefaultSource{}
}

// Load applies default values to configuration.
func (s *DefaultSource) Load(cfg interface{}) error {
	return s.loadStruct(cfg)
}

// loadStruct recursively applies default values to a struct.
func (s *DefaultSource) loadStruct(cfg interface{}) error {
	rv := reflect.ValueOf(cfg)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct or pointer to struct")
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get config tag
		tag := field.Tag.Get("config")
		options := parseTagOptions(tag)

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := s.loadStruct(fieldValue.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// Skip if field already has a value
		if !isZeroValue(fieldValue) {
			continue
		}

		// Apply default value if specified
		if defaultValue := options["default"]; defaultValue != "" {
			if err := s.setFieldValue(fieldValue, defaultValue); err != nil {
				return fmt.Errorf("field %q: %w", field.Name, err)
			}
		}
	}

	return nil
}

// setFieldValue sets a field value from a string (same as EnvSource).
func (s *DefaultSource) setFieldValue(fieldValue reflect.Value, value string) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer: %w", err)
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %w", err)
		}
		fieldValue.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %w", err)
		}
		fieldValue.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	return nil
}

// isZeroValue checks if a value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return false
	}
}

