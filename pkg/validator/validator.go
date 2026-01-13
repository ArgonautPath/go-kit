package validator

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidateStruct validates a struct using struct tags.
// It recursively validates nested structs and collects all validation errors.
//
// Example:
//
//	type User struct {
//	    Email string `validate:"required,email"`
//	    Age   int    `validate:"required,range=18,120"`
//	}
//
//	user := User{Email: "invalid", Age: 15}
//	err := ValidateStruct(user)
//	if err != nil {
//	    // Handle validation errors
//	}
func ValidateStruct(v interface{}) error {
	return ValidateStructWithTag(v, DefaultTagName)
}

// ValidateStructWithTag validates a struct using a custom struct tag name.
func ValidateStructWithTag(v interface{}, tagName string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct requires a struct or pointer to struct")
	}

	var errors ValidationErrors
	validateStructFields(rv, "", tagName, &errors)

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateStructFields iterates through struct fields and validates them.
func validateStructFields(rv reflect.Value, fieldPath string, tagName string, errors *ValidationErrors) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Build field path for nested structs
		currentPath := field.Name
		if fieldPath != "" {
			currentPath = fieldPath + "." + field.Name
		}

		if err := validateField(field, fieldValue, currentPath, tagName, errors); err != nil {
			errors.Add(err)
		}
	}
}

// validateField validates a single struct field.
func validateField(field reflect.StructField, fieldValue reflect.Value, fieldPath, tagName string, errors *ValidationErrors) *ValidationError {
	// Skip unexported fields
	if !fieldValue.CanInterface() {
		return nil
	}

	tag := field.Tag.Get(tagName)
	
	// Check if this is a nested struct that should be validated recursively
	isNestedStruct := (fieldValue.Kind() == reflect.Struct) || 
		(fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() && fieldValue.Elem().Kind() == reflect.Struct)
	
	if tag == "" {
		// No tag - recursively validate nested structs
		if isNestedStruct {
			return validateNestedStruct(field, fieldValue, fieldPath, tagName, errors)
		}
		return nil
	}

	// Parse tag into individual rules
	// Format: "required,email,range=1,100" -> ["required", "email", "range=1,100"]
	rules := splitTagIntoParts(tag)

	// Check if field is required
	hasRequired := false
	for _, rule := range rules {
		if strings.TrimSpace(rule) == "required" {
			hasRequired = true
			if err := validateRequired(field, fieldValue, fieldPath); err != nil {
				return err
			}
			break
		}
	}

	// Validate all rules (except required, which is already handled)
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" || rule == "required" {
			continue // Already handled required
		}

		// Skip validation if value is empty and field is not required
		if shouldSkipValidation(fieldValue, rule, hasRequired) {
			continue
		}

		if err := executeRule(rule, fieldValue.Interface()); err != nil {
			return &ValidationError{
				FieldPath: fieldPath,
				Field:     field.Name,
				Value:     fieldValue.Interface(),
				Message:   err.Error(),
				Rule:      rule,
			}
		}
	}

	// After validating the field's own rules, validate nested struct if it exists
	if isNestedStruct {
		validateNestedStruct(field, fieldValue, fieldPath, tagName, errors)
	}

	return nil
}

// validateNestedStruct recursively validates nested structs.
func validateNestedStruct(field reflect.StructField, fieldValue reflect.Value, fieldPath, tagName string, errors *ValidationErrors) *ValidationError {
	if fieldValue.Kind() == reflect.Struct {
		validateStructFields(fieldValue, fieldPath, tagName, errors)
	} else if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
		if fieldValue.Elem().Kind() == reflect.Struct {
			validateStructFields(fieldValue.Elem(), fieldPath, tagName, errors)
		}
	} else if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
		// Nil pointer to struct - check if it's required
		tag := field.Tag.Get(tagName)
		if tag != "" {
			options := parseTagOptions(tag)
			if _, isRequired := options["required"]; isRequired {
				return &ValidationError{
					FieldPath: fieldPath,
					Field:     field.Name,
					Value:     nil,
					Message:   "required field is nil",
					Rule:      "required",
				}
			}
		}
	}
	return nil
}

// validateRequired checks if a required field is set.
func validateRequired(field reflect.StructField, fieldValue reflect.Value, fieldPath string) *ValidationError {
	validator := NewRequiredValidator()
	if err := validator.Validate(fieldValue.Interface()); err != nil {
		return &ValidationError{
			FieldPath: fieldPath,
			Field:     field.Name,
			Value:     fieldValue.Interface(),
			Message:   err.Error(),
			Rule:      "required",
		}
	}
	return nil
}
