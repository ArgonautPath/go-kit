package validator

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// Validator defines the interface for validating values.
type Validator interface {
	Validate(value interface{}) error
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
	// RFC 5322 compliant email regex (simplified)
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
		return fmt.Errorf("url validator requires string value")
	}

	if str == "" {
		return fmt.Errorf("url cannot be empty")
	}

	_, err := url.ParseRequestURI(str)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	return nil
}

// rangeValidator validates numeric values within a range.
type rangeValidator struct {
	min *float64
	max *float64
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

// lengthValidator validates string/slice length.
type lengthValidator struct {
	min   *int
	max   *int
	exact *int
}

// NewLengthValidator creates a new length validator.
func NewLengthValidator(min, max, exact *int) Validator {
	return &lengthValidator{min: min, max: max, exact: exact}
}

// Validate checks if the string/slice length is within the specified range.
func (v *lengthValidator) Validate(value interface{}) error {
	var length int

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		length = len(rv.String())
	case reflect.Slice, reflect.Array, reflect.Map:
		length = rv.Len()
	default:
		return fmt.Errorf("length validator requires string, slice, array, or map value")
	}

	if v.exact != nil {
		if length != *v.exact {
			return fmt.Errorf("length must be exactly %d, got %d", *v.exact, length)
		}
		return nil
	}

	if v.min != nil && length < *v.min {
		return fmt.Errorf("length %d is less than minimum %d", length, *v.min)
	}

	if v.max != nil && length > *v.max {
		return fmt.Errorf("length %d is greater than maximum %d", length, *v.max)
	}

	return nil
}

// minValidator validates minimum value/length.
type minValidator struct {
	min float64
}

// NewMinValidator creates a new min validator.
func NewMinValidator(min float64) Validator {
	return &minValidator{min: min}
}

// Validate checks if the value is greater than or equal to the minimum.
func (v *minValidator) Validate(value interface{}) error {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		length := rv.Len()
		if float64(length) < v.min {
			return fmt.Errorf("length %d is less than minimum %v", length, v.min)
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(rv.Int()) < v.min {
			return fmt.Errorf("value %v is less than minimum %v", rv.Int(), v.min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(rv.Uint()) < v.min {
			return fmt.Errorf("value %v is less than minimum %v", rv.Uint(), v.min)
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() < v.min {
			return fmt.Errorf("value %v is less than minimum %v", rv.Float(), v.min)
		}
	default:
		return fmt.Errorf("min validator requires numeric, string, slice, array, or map value")
	}

	return nil
}

// maxValidator validates maximum value/length.
type maxValidator struct {
	max float64
}

// NewMaxValidator creates a new max validator.
func NewMaxValidator(max float64) Validator {
	return &maxValidator{max: max}
}

// Validate checks if the value is less than or equal to the maximum.
func (v *maxValidator) Validate(value interface{}) error {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		length := rv.Len()
		if float64(length) > v.max {
			return fmt.Errorf("length %d is greater than maximum %v", length, v.max)
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(rv.Int()) > v.max {
			return fmt.Errorf("value %v is greater than maximum %v", rv.Int(), v.max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(rv.Uint()) > v.max {
			return fmt.Errorf("value %v is greater than maximum %v", rv.Uint(), v.max)
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() > v.max {
			return fmt.Errorf("value %v is greater than maximum %v", rv.Float(), v.max)
		}
	default:
		return fmt.Errorf("max validator requires numeric, string, slice, array, or map value")
	}

	return nil
}

// regexValidator validates string against a regex pattern.
type regexValidator struct {
	pattern *regexp.Regexp
}

// NewRegexValidator creates a new regex validator.
func NewRegexValidator(pattern string) (Validator, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}
	return &regexValidator{pattern: re}, nil
}

// Validate checks if the value matches the regex pattern.
func (v *regexValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("regex validator requires string value")
	}

	if !v.pattern.MatchString(str) {
		return fmt.Errorf("value does not match pattern")
	}

	return nil
}

// oneOfValidator validates that value is one of allowed values.
type oneOfValidator struct {
	allowed []string
}

// NewOneOfValidator creates a new oneOf validator.
func NewOneOfValidator(allowed []string) Validator {
	return &oneOfValidator{allowed: allowed}
}

// Validate checks if the value is one of the allowed values.
func (v *oneOfValidator) Validate(value interface{}) error {
	var str string
	switch val := value.(type) {
	case string:
		str = val
	case fmt.Stringer:
		str = val.String()
	default:
		str = fmt.Sprintf("%v", value)
	}

	for _, allowed := range v.allowed {
		if str == allowed {
			return nil
		}
	}

	return fmt.Errorf("value %q is not one of allowed values: %v", str, v.allowed)
}

// alphaValidator validates that string contains only letters.
type alphaValidator struct {
	pattern *regexp.Regexp
}

// NewAlphaValidator creates a new alpha validator.
func NewAlphaValidator() Validator {
	pattern := regexp.MustCompile(`^[a-zA-Z]+$`)
	return &alphaValidator{pattern: pattern}
}

// Validate checks if the value contains only letters.
func (v *alphaValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("alpha validator requires string value")
	}

	if !v.pattern.MatchString(str) {
		return fmt.Errorf("value must contain only letters")
	}

	return nil
}

// alphanumericValidator validates that string contains only letters and numbers.
type alphanumericValidator struct {
	pattern *regexp.Regexp
}

// NewAlphanumericValidator creates a new alphanumeric validator.
func NewAlphanumericValidator() Validator {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return &alphanumericValidator{pattern: pattern}
}

// Validate checks if the value contains only letters and numbers.
func (v *alphanumericValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("alphanumeric validator requires string value")
	}

	if !v.pattern.MatchString(str) {
		return fmt.Errorf("value must contain only letters and numbers")
	}

	return nil
}

// numericValidator validates that string contains only numbers.
type numericValidator struct {
	pattern *regexp.Regexp
}

// NewNumericValidator creates a new numeric validator.
func NewNumericValidator() Validator {
	pattern := regexp.MustCompile(`^[0-9]+$`)
	return &numericValidator{pattern: pattern}
}

// Validate checks if the value contains only numbers.
func (v *numericValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("numeric validator requires string value")
	}

	if !v.pattern.MatchString(str) {
		return fmt.Errorf("value must contain only numbers")
	}

	return nil
}
