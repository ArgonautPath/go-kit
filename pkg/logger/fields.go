package logger

import (
	"time"
)

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a field with a string value.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates a field with an integer value.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates a field with an int64 value.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a field with a float64 value.
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a field with a boolean value.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error creates a field from an error.
func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Duration creates a field with a time.Duration value.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// Time creates a field with a time.Time value.
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

// Any creates a field with any value.
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Fields converts a map to a slice of Field.
func Fields(fields map[string]interface{}) []Field {
	result := make([]Field, 0, len(fields))
	for k, v := range fields {
		result = append(result, Field{Key: k, Value: v})
	}
	return result
}

