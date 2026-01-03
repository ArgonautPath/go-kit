package config

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator defines the interface for validating configuration values.
type Validator interface {
	Validate(value interface{}) error
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %q: %s (value: %v)", e.Field, e.Message, e.Value)
}

// requiredValidator validates that a value is not empty.
type requiredValidator struct{}

// NewRequiredValidator creates a new required validator.
func NewRequiredValidator() Validator {
	return &requiredValidator{}
}

// Validate checks if the value is not empty.
func (v *requiredValidator) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf("required field is nil")
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		if strings.TrimSpace(rv.String()) == "" {
			return fmt.Errorf("required field is empty")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() == 0 {
			return fmt.Errorf("required field is zero")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if rv.Uint() == 0 {
			return fmt.Errorf("required field is zero")
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() == 0 {
			return fmt.Errorf("required field is zero")
		}
	case reflect.Bool:
		// Bool can't be "required" in the traditional sense, but we check anyway
		// This is a design choice - you might want to allow false as valid
	case reflect.Slice, reflect.Map, reflect.Array:
		if rv.Len() == 0 {
			return fmt.Errorf("required field is empty")
		}
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return fmt.Errorf("required field is nil")
		}
	}

	return nil
}

// emailValidator validates email format.
type emailValidator struct {
	pattern *regexp.Regexp
}

// NewEmailValidator creates a new email validator.
func NewEmailValidator() Validator {
	// Simple email regex pattern
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return &emailValidator{pattern: pattern}
}

// Validate checks if the value is a valid email address.
func (v *emailValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("email validator requires string value")
	}

	if !v.pattern.MatchString(str) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// urlValidator validates URL format.
type urlValidator struct{}

// NewURLValidator creates a new URL validator.
func NewURLValidator() Validator {
	return &urlValidator{}
}

// Validate checks if the value is a valid URL.
func (v *urlValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("URL validator requires string value")
	}

	parsedURL, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if URL has a scheme and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: missing scheme or host")
	}

	return nil
}

// rangeValidator validates numeric values within a range.
type rangeValidator struct {
	min, max *float64
}

// NewRangeValidator creates a new range validator.
func NewRangeValidator(min, max *float64) Validator {
	return &rangeValidator{min: min, max: max}
}

// Validate checks if the numeric value is within the specified range.
func (v *rangeValidator) Validate(value interface{}) error {
	var num float64

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num = float64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num = float64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		num = rv.Float()
	default:
		return fmt.Errorf("range validator requires numeric value")
	}

	if v.min != nil && num < *v.min {
		return fmt.Errorf("value %v is less than minimum %v", num, *v.min)
	}

	if v.max != nil && num > *v.max {
		return fmt.Errorf("value %v is greater than maximum %v", num, *v.max)
	}

	return nil
}

// ValidateStruct validates a struct using struct tags.
func ValidateStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct requires a struct or pointer to struct")
	}

	return validateStructFields(rv)
}

// validateStructFields iterates through struct fields and validates them.
func validateStructFields(rv reflect.Value) error {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if err := validateField(field, fieldValue); err != nil {
			return err
		}
	}
	return nil
}

// validateField validates a single struct field.
func validateField(field reflect.StructField, fieldValue reflect.Value) error {
	// Skip unexported fields
	if !fieldValue.CanInterface() {
		return nil
	}

	tag := field.Tag.Get("config")
	if tag == "" {
		// Recursively validate nested structs
		return validateNestedStruct(field, fieldValue)
	}

	options := parseTagOptions(tag)

	// Validate required field
	if err := validateRequired(field, fieldValue, options); err != nil {
		return err
	}

	// Validate custom rules
	if err := validateRules(field, fieldValue, options); err != nil {
		return err
	}

	return nil
}

// validateNestedStruct recursively validates nested structs.
func validateNestedStruct(field reflect.StructField, fieldValue reflect.Value) error {
	if fieldValue.Kind() == reflect.Struct {
		if err := ValidateStruct(fieldValue.Interface()); err != nil {
			return fmt.Errorf("field %q: %w", field.Name, err)
		}
	}
	return nil
}

// validateRequired checks if a required field is set.
func validateRequired(field reflect.StructField, fieldValue reflect.Value, options map[string]string) error {
	if _, isRequired := options["required"]; !isRequired {
		return nil
	}

	validator := NewRequiredValidator()
	if err := validator.Validate(fieldValue.Interface()); err != nil {
		return &ValidationError{
			Field:   field.Name,
			Value:   fieldValue.Interface(),
			Message: err.Error(),
		}
	}
	return nil
}

// validateRules applies custom validation rules to a field.
func validateRules(field reflect.StructField, fieldValue reflect.Value, options map[string]string) error {
	validateRule := options["validate"]
	if validateRule == "" {
		return nil
	}

	if shouldSkipValidation(fieldValue, validateRule, options) {
		return nil
	}

	if err := validateByRule(validateRule, fieldValue.Interface()); err != nil {
		return &ValidationError{
			Field:   field.Name,
			Value:   fieldValue.Interface(),
			Message: err.Error(),
		}
	}
	return nil
}

// shouldSkipValidation determines if validation should be skipped for a field.
// Range validation always runs (even for zero values), but other validations
// are skipped if the value is empty and the field is not required.
func shouldSkipValidation(fieldValue reflect.Value, validateRule string, options map[string]string) bool {
	// Always validate range, even for zero values
	if strings.HasPrefix(validateRule, "range=") {
		return false
	}

	// Skip validation if value is empty and field is not required
	_, isRequired := options["required"]
	return isZeroValue(fieldValue) && !isRequired
}

// parseTagOptions parses struct tag options.
// Handles commas inside values (e.g., "validate=range=1,65535").
func parseTagOptions(tag string) map[string]string {
	options := make(map[string]string)
	parts := splitTagIntoParts(tag)
	mergedParts := mergeRangeValueParts(parts)

	for _, part := range mergedParts {
		key, value := parseKeyValuePair(part)
		if key != "" {
			options[key] = value
		}
	}

	return options
}

// splitTagIntoParts splits a tag string by commas, preserving whitespace for trimming.
func splitTagIntoParts(tag string) []string {
	parts := strings.Split(tag, ",")
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed = append(trimmed, strings.TrimSpace(part))
	}
	return trimmed
}

// mergeRangeValueParts merges parts that belong to range values.
// Range values contain commas (e.g., "range=1,65535"), so we need to merge
// parts that were incorrectly split by the comma separator.
func mergeRangeValueParts(parts []string) []string {
	if len(parts) == 0 {
		return parts
	}

	merged := make([]string, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		// Check if this part should be merged with the previous one
		if i > 0 && len(merged) > 0 {
			prevPart := merged[len(merged)-1]
			if isRangeValueContinuation(prevPart, part) {
				merged[len(merged)-1] = prevPart + "," + part
				continue
			}
		}

		merged = append(merged, part)
	}

	return merged
}

// isRangeValueContinuation checks if a part should be merged with the previous part.
// This happens when the previous part contains "range=" but doesn't have a complete
// range value yet (missing the comma-separated min,max values).
func isRangeValueContinuation(prevPart, currentPart string) bool {
	// Previous part must contain "range="
	if !strings.Contains(prevPart, "range=") {
		return false
	}

	// If previous part already has a comma, it's complete
	if strings.Contains(prevPart, ",") && !strings.HasSuffix(prevPart, "range=") {
		return false
	}

	// Current part should look like a number (range value continuation)
	// This is a heuristic: if it starts with a digit or minus sign, it's likely a range value
	trimmed := strings.TrimSpace(currentPart)
	if len(trimmed) > 0 {
		firstChar := trimmed[0]
		if (firstChar >= '0' && firstChar <= '9') || firstChar == '-' {
			return true
		}
	}

	return false
}

// parseKeyValuePair parses a single key-value pair from a tag part.
// Returns the key and value. If no "=" is present, the entire part is the key
// and the value is empty (for boolean flags like "required").
func parseKeyValuePair(part string) (key, value string) {
	part = strings.TrimSpace(part)
	if part == "" {
		return "", ""
	}

	if !strings.Contains(part, "=") {
		return part, ""
	}

	kv := strings.SplitN(part, "=", 2)
	if len(kv) != 2 {
		return part, ""
	}

	key = strings.TrimSpace(kv[0])
	value = strings.TrimSpace(kv[1])
	return key, value
}

// validateByRule validates a value using a validation rule.
func validateByRule(rule string, value interface{}) error {
	ruleName, ruleValue := parseRule(rule)

	switch ruleName {
	case "email":
		return validateEmail(value)
	case "url":
		return validateURL(value)
	case "range":
		return validateRange(ruleValue, value)
	default:
		return fmt.Errorf("unknown validation rule: %s", ruleName)
	}
}

// parseRule parses a validation rule string into rule name and value.
// Example: "range=1,100" -> ("range", "1,100")
func parseRule(rule string) (ruleName, ruleValue string) {
	parts := strings.SplitN(rule, "=", 2)
	ruleName = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		ruleValue = strings.TrimSpace(parts[1])
	}
	return ruleName, ruleValue
}

// validateEmail validates an email address.
func validateEmail(value interface{}) error {
	return NewEmailValidator().Validate(value)
}

// validateURL validates a URL.
func validateURL(value interface{}) error {
	return NewURLValidator().Validate(value)
}

// validateRange validates a numeric value against a range specification.
// ruleValue format: "min,max" or "min," or ",max"
func validateRange(ruleValue string, value interface{}) error {
	if ruleValue == "" {
		return fmt.Errorf("range validator requires min and max values")
	}

	min, max, err := parseRangeValues(ruleValue)
	if err != nil {
		return err
	}

	if min == nil && max == nil {
		return fmt.Errorf("range validator requires at least min or max value")
	}

	return NewRangeValidator(min, max).Validate(value)
}

// parseRangeValues parses range values from a string like "1,100" or "1," or ",100".
func parseRangeValues(ruleValue string) (min, max *float64, err error) {
	rangeParts := strings.Split(ruleValue, ",")

	// Parse minimum value
	if len(rangeParts) > 0 && rangeParts[0] != "" {
		val, parseErr := strconv.ParseFloat(strings.TrimSpace(rangeParts[0]), 64)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("invalid min value in range: %w", parseErr)
		}
		min = &val
	}

	// Parse maximum value
	if len(rangeParts) > 1 && rangeParts[1] != "" {
		val, parseErr := strconv.ParseFloat(strings.TrimSpace(rangeParts[1]), 64)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("invalid max value in range: %w", parseErr)
		}
		max = &val
	}

	return min, max, nil
}
