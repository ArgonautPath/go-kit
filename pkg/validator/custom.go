package validator

import (
	"fmt"
	"sync"
)

// CustomValidatorFunc is a function type for custom validators.
type CustomValidatorFunc func(value interface{}) error

var (
	// customValidators is a registry of custom validators.
	customValidators = make(map[string]CustomValidatorFunc)
	// customValidatorsMu protects the customValidators map.
	customValidatorsMu sync.RWMutex
)

// RegisterValidator registers a custom validator function.
// The name must be unique and will be used in struct tags.
// If a validator with the same name already exists, it will be overwritten.
//
// Example:
//
//	validator.RegisterValidator("custom", func(value interface{}) error {
//	    str, ok := value.(string)
//	    if !ok {
//	        return fmt.Errorf("expected string")
//	    }
//	    if str != "valid" {
//	        return fmt.Errorf("value must be 'valid'")
//	    }
//	    return nil
//	})
//
//	type Config struct {
//	    Field string `validate:"required,custom"`
//	}
func RegisterValidator(name string, fn CustomValidatorFunc) error {
	if name == "" {
		return fmt.Errorf("validator name cannot be empty")
	}

	if fn == nil {
		return fmt.Errorf("validator function cannot be nil")
	}

	// Check if name conflicts with built-in validators
	builtinNames := []string{
		"required", "email", "url", "range", "len", "min", "max",
		"regex", "oneof", "alpha", "alphanumeric", "numeric",
	}
	for _, builtin := range builtinNames {
		if name == builtin {
			return fmt.Errorf("validator name %q conflicts with built-in validator", name)
		}
	}

	customValidatorsMu.Lock()
	defer customValidatorsMu.Unlock()

	customValidators[name] = fn
	return nil
}

// UnregisterValidator removes a custom validator.
func UnregisterValidator(name string) {
	customValidatorsMu.Lock()
	defer customValidatorsMu.Unlock()

	delete(customValidators, name)
}

// getCustomValidator retrieves a custom validator by name.
func getCustomValidator(name string) (CustomValidatorFunc, bool) {
	customValidatorsMu.RLock()
	defer customValidatorsMu.RUnlock()

	validator, ok := customValidators[name]
	return validator, ok
}

// ListCustomValidators returns a list of all registered custom validator names.
func ListCustomValidators() []string {
	customValidatorsMu.RLock()
	defer customValidatorsMu.RUnlock()

	names := make([]string, 0, len(customValidators))
	for name := range customValidators {
		names = append(names, name)
	}
	return names
}
