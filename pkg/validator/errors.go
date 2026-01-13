package validator

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error for a field.
type ValidationError struct {
	// FieldPath is the full path to the field (e.g., "User.Email", "Config.Database.Host").
	FieldPath string
	// Field is the field name.
	Field string
	// Value is the value that failed validation.
	Value interface{}
	// Message is the error message.
	Message string
	// Rule is the validation rule that failed (e.g., "required", "email", "range=1,100").
	Rule string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.FieldPath != "" && e.FieldPath != e.Field {
		return fmt.Sprintf("validation error for field %q: %s (value: %v, rule: %s)", e.FieldPath, e.Message, e.Value, e.Rule)
	}
	return fmt.Sprintf("validation error for field %q: %s (value: %v, rule: %s)", e.Field, e.Message, e.Value, e.Rule)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	if len(e) == 1 {
		return e[0].Error()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d validation errors:\n", len(e)))
	for i, err := range e {
		b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return b.String()
}

// Errors returns all validation errors.
func (e ValidationErrors) Errors() []*ValidationError {
	return []*ValidationError(e)
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// First returns the first validation error, or nil if there are no errors.
func (e ValidationErrors) First() *ValidationError {
	if len(e) == 0 {
		return nil
	}
	return e[0]
}

// Add adds a validation error to the collection.
func (e *ValidationErrors) Add(err *ValidationError) {
	*e = append(*e, err)
}

// AddAll adds all validation errors from another collection.
func (e *ValidationErrors) AddAll(errs ValidationErrors) {
	*e = append(*e, errs...)
}

// IsValidationError checks if an error is a ValidationError or ValidationErrors.
func IsValidationError(err error) bool {
	_, ok1 := err.(*ValidationError)
	_, ok2 := err.(ValidationErrors)
	return ok1 || ok2
}

// AsValidationErrors converts an error to ValidationErrors if possible.
func AsValidationErrors(err error) (ValidationErrors, bool) {
	switch v := err.(type) {
	case ValidationErrors:
		return v, true
	case *ValidationError:
		return ValidationErrors{v}, true
	default:
		return nil, false
	}
}
