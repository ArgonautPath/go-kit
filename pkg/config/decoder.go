package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Decoder defines the interface for decoding configuration files.
type Decoder interface {
	Decode(r io.Reader, v interface{}) error
}

// Format represents the configuration file format.
type Format string

const (
	// YAMLFormat represents YAML file format.
	YAMLFormat Format = "yaml"
	// JSONFormat represents JSON file format.
	JSONFormat Format = "json"
	// UnknownFormat represents an unknown or unsupported format.
	UnknownFormat Format = "unknown"
)

// yamlDecoder decodes YAML files.
type yamlDecoder struct{}

// NewYAMLDecoder creates a new YAML decoder.
func NewYAMLDecoder() Decoder {
	return &yamlDecoder{}
}

// Decode decodes YAML data from the reader into v.
func (d *yamlDecoder) Decode(r io.Reader, v interface{}) error {
	return yaml.NewDecoder(r).Decode(v)
}

// jsonDecoder decodes JSON files.
type jsonDecoder struct{}

// NewJSONDecoder creates a new JSON decoder.
func NewJSONDecoder() Decoder {
	return &jsonDecoder{}
}

// Decode decodes JSON data from the reader into v.
func (d *jsonDecoder) Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// DetectFormat detects the file format from the file extension.
func DetectFormat(filename string) Format {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "yaml", "yml":
		return YAMLFormat
	case "json":
		return JSONFormat
	default:
		return UnknownFormat
	}
}

// NewDecoder creates a new decoder based on the file format.
func NewDecoder(format Format) (Decoder, error) {
	switch format {
	case YAMLFormat:
		return NewYAMLDecoder(), nil
	case JSONFormat:
		return NewJSONDecoder(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// DecodeFile decodes a configuration file into v.
// The file format is automatically detected from the extension.
func DecodeFile(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	format := DetectFormat(path)
	if format == UnknownFormat {
		return fmt.Errorf("unknown file format: %s", path)
	}

	decoder, err := NewDecoder(format)
	if err != nil {
		return fmt.Errorf("create decoder: %w", err)
	}

	if err := decoder.Decode(file, v); err != nil {
		return fmt.Errorf("decode file: %w", err)
	}

	return nil
}

