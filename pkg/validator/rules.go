package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// parseRule parses a validation rule string into rule name and value.
// Example: "range=1,100" -> ("range", "1,100")
// Example: "required" -> ("required", "")
func parseRule(rule string) (ruleName, ruleValue string) {
	parts := strings.SplitN(rule, "=", 2)
	ruleName = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		ruleValue = strings.TrimSpace(parts[1])
	}
	return ruleName, ruleValue
}

// executeRule executes a validation rule on a value.
func executeRule(rule string, value interface{}) error {
	ruleName, ruleValue := parseRule(rule)

	switch ruleName {
	case "required":
		return NewRequiredValidator().Validate(value)
	case "email":
		return NewEmailValidator().Validate(value)
	case "url":
		return NewURLValidator().Validate(value)
	case "alpha":
		return NewAlphaValidator().Validate(value)
	case "alphanumeric":
		return NewAlphanumericValidator().Validate(value)
	case "numeric":
		return NewNumericValidator().Validate(value)
	case "range":
		return executeRangeRule(ruleValue, value)
	case "len":
		return executeLengthRule(ruleValue, value)
	case "min":
		return executeMinRule(ruleValue, value)
	case "max":
		return executeMaxRule(ruleValue, value)
	case "regex":
		return executeRegexRule(ruleValue, value)
	case "oneof":
		return executeOneOfRule(ruleValue, value)
	default:
		// Try custom validator
		if customValidator, ok := getCustomValidator(ruleName); ok {
			return customValidator(value)
		}
		return fmt.Errorf("unknown validation rule: %s", ruleName)
	}
}

// executeRangeRule executes a range validation rule.
// Format: "range=min,max" or "range=,max" or "range=min,"
func executeRangeRule(ruleValue string, value interface{}) error {
	var min, max *float64

	if ruleValue != "" {
		parts := strings.Split(ruleValue, ",")
		if len(parts) != 2 {
			return fmt.Errorf("invalid range format: expected 'min,max'")
		}

		if parts[0] != "" {
			minVal, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return fmt.Errorf("invalid range minimum: %w", err)
			}
			min = &minVal
		}

		if parts[1] != "" {
			maxVal, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return fmt.Errorf("invalid range maximum: %w", err)
			}
			max = &maxVal
		}
	}

	return NewRangeValidator(min, max).Validate(value)
}

// executeLengthRule executes a length validation rule.
// Format: "len=exact" or "len=min,max"
func executeLengthRule(ruleValue string, value interface{}) error {
	if ruleValue == "" {
		return fmt.Errorf("length rule requires a value")
	}

	parts := strings.Split(ruleValue, ",")
	if len(parts) == 1 {
		// Exact length
		exact, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return fmt.Errorf("invalid length value: %w", err)
		}
		return NewLengthValidator(nil, nil, &exact).Validate(value)
	}

	if len(parts) != 2 {
		return fmt.Errorf("invalid length format: expected 'exact' or 'min,max'")
	}

	var min, max *int

	if parts[0] != "" {
		minVal, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return fmt.Errorf("invalid length minimum: %w", err)
		}
		min = &minVal
	}

	if parts[1] != "" {
		maxVal, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return fmt.Errorf("invalid length maximum: %w", err)
		}
		max = &maxVal
	}

	return NewLengthValidator(min, max, nil).Validate(value)
}

// executeMinRule executes a min validation rule.
func executeMinRule(ruleValue string, value interface{}) error {
	if ruleValue == "" {
		return fmt.Errorf("min rule requires a value")
	}

	min, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return fmt.Errorf("invalid min value: %w", err)
	}

	return NewMinValidator(min).Validate(value)
}

// executeMaxRule executes a max validation rule.
func executeMaxRule(ruleValue string, value interface{}) error {
	if ruleValue == "" {
		return fmt.Errorf("max rule requires a value")
	}

	max, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return fmt.Errorf("invalid max value: %w", err)
	}

	return NewMaxValidator(max).Validate(value)
}

// executeRegexRule executes a regex validation rule.
func executeRegexRule(ruleValue string, value interface{}) error {
	if ruleValue == "" {
		return fmt.Errorf("regex rule requires a pattern")
	}

	validator, err := NewRegexValidator(ruleValue)
	if err != nil {
		return err
	}

	return validator.Validate(value)
}

// executeOneOfRule executes a oneOf validation rule.
func executeOneOfRule(ruleValue string, value interface{}) error {
	if ruleValue == "" {
		return fmt.Errorf("oneof rule requires allowed values")
	}

	allowed := strings.Split(ruleValue, ",")
	for i := range allowed {
		allowed[i] = strings.TrimSpace(allowed[i])
	}

	return NewOneOfValidator(allowed).Validate(value)
}

// shouldSkipValidation determines if validation should be skipped for a field.
// Range validation always runs (even for zero values), but other validations
// are skipped if the value is empty and the field is not required.
func shouldSkipValidation(fieldValue reflect.Value, rule string, hasRequired bool) bool {
	// Always validate range, even for zero values
	if strings.HasPrefix(rule, "range=") {
		return false
	}

	// Skip validation if value is empty and field is not required
	return isZeroValue(fieldValue) && !hasRequired
}

// isZeroValue checks if a reflect.Value represents a zero value.
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
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}
