package validator

import (
	"errors"
	"testing"
)

func TestRequiredValidator(t *testing.T) {
	validator := NewRequiredValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid string", "test", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"valid int", 42, false},
		{"zero int", 0, true},
		{"valid slice", []string{"a"}, false},
		{"empty slice", []string{}, true},
		{"nil value", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	validator := NewEmailValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"invalid email", "invalid", true},
		{"invalid format", "notanemail", true},
		{"missing @", "testexample.com", true},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestURLValidator(t *testing.T) {
	validator := NewURLValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid http url", "http://example.com", false},
		{"valid https url", "https://example.com", false},
		{"invalid url", "not a url", true},
		{"empty url", "", true},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRangeValidator(t *testing.T) {
	tests := []struct {
		name    string
		min     *float64
		max     *float64
		value   interface{}
		wantErr bool
	}{
		{"in range", floatPtr(1), floatPtr(100), 50, false},
		{"below min", floatPtr(10), floatPtr(100), 5, true},
		{"above max", floatPtr(1), floatPtr(100), 150, true},
		{"only min", floatPtr(10), nil, 15, false},
		{"only max", nil, floatPtr(100), 50, false},
		{"non-numeric", floatPtr(1), floatPtr(100), "string", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewRangeValidator(tt.min, tt.max)
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLengthValidator(t *testing.T) {
	tests := []struct {
		name    string
		min     *int
		max     *int
		exact   *int
		value   interface{}
		wantErr bool
	}{
		{"exact length", nil, nil, intPtr(5), "hello", false},
		{"wrong exact length", nil, nil, intPtr(5), "hi", true},
		{"in range", intPtr(2), intPtr(10), nil, "hello", false},
		{"below min", intPtr(5), intPtr(10), nil, "hi", true},
		{"above max", intPtr(2), intPtr(5), nil, "hello world", true},
		{"slice length", nil, nil, intPtr(3), []int{1, 2, 3}, false},
		{"non-string/slice", nil, nil, intPtr(5), 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewLengthValidator(tt.min, tt.max, tt.exact)
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMinValidator(t *testing.T) {
	tests := []struct {
		name    string
		min     float64
		value   interface{}
		wantErr bool
	}{
		{"valid number", 10, 15, false},
		{"below min", 10, 5, true},
		{"valid string length", 5, "hello world", false},
		{"short string", 10, "hi", true},
		{"valid slice length", 3, []int{1, 2, 3, 4}, false},
		{"short slice", 5, []int{1, 2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewMinValidator(tt.min)
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaxValidator(t *testing.T) {
	tests := []struct {
		name    string
		max     float64
		value   interface{}
		wantErr bool
	}{
		{"valid number", 100, 50, false},
		{"above max", 100, 150, true},
		{"valid string length", 10, "hello", false},
		{"long string", 5, "hello world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewMaxValidator(tt.max)
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegexValidator(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		value   interface{}
		wantErr bool
	}{
		{"matches pattern", "^[a-z]+$", "hello", false},
		{"doesn't match", "^[a-z]+$", "Hello123", true},
		{"invalid pattern", "[invalid", "test", true},
		{"non-string", "^[a-z]+$", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewRegexValidator(tt.pattern)
			if err != nil && !tt.wantErr {
				t.Fatalf("NewRegexValidator() error = %v", err)
			}
			if validator != nil {
				err = validator.Validate(tt.value)
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestOneOfValidator(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		value   interface{}
		wantErr bool
	}{
		{"valid value", []string{"red", "green", "blue"}, "red", false},
		{"invalid value", []string{"red", "green", "blue"}, "yellow", true},
		{"stringer interface", []string{"1", "2", "3"}, testStringer("2"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewOneOfValidator(tt.allowed)
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlphaValidator(t *testing.T) {
	validator := NewAlphaValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"only letters", "Hello", false},
		{"with numbers", "Hello123", true},
		{"with spaces", "Hello World", true},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlphanumericValidator(t *testing.T) {
	validator := NewAlphanumericValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"letters and numbers", "Hello123", false},
		{"only letters", "Hello", false},
		{"only numbers", "123", false},
		{"with spaces", "Hello 123", true},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNumericValidator(t *testing.T) {
	validator := NewNumericValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"only numbers", "123", false},
		{"with letters", "123abc", true},
		{"with spaces", "123 456", true},
		{"non-string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	type User struct {
		Email    string `validate:"required,email"`
		Age      int    `validate:"required,range=18,120"`
		Password string `validate:"required,min=8"`
	}

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    User{Email: "test@example.com", Age: 25, Password: "password123"},
			wantErr: false,
		},
		{
			name:    "missing email",
			user:    User{Age: 25, Password: "password123"},
			wantErr: true,
		},
		{
			name:    "invalid email",
			user:    User{Email: "invalid", Age: 25, Password: "password123"},
			wantErr: true,
		},
		{
			name:    "age too low",
			user:    User{Email: "test@example.com", Age: 15, Password: "password123"},
			wantErr: true,
		},
		{
			name:    "password too short",
			user:    User{Email: "test@example.com", Age: 25, Password: "short"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct_Nested(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type User struct {
		Name    string  `validate:"required"`
		Address Address `validate:"required"`
	}

	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name:    "valid nested",
			user:    User{Name: "John", Address: Address{Street: "123 Main", City: "NYC"}},
			wantErr: false,
		},
		{
			name:    "missing name",
			user:    User{Address: Address{Street: "123 Main", City: "NYC"}},
			wantErr: true,
		},
		{
			name:    "missing nested field",
			user:    User{Name: "John", Address: Address{Street: "123 Main"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	errs := ValidationErrors{
		&ValidationError{Field: "Email", Message: "invalid email"},
		&ValidationError{Field: "Age", Message: "too young"},
	}

	if !errs.HasErrors() {
		t.Error("HasErrors() = false, want true")
	}

	if errs.First() == nil {
		t.Error("First() = nil, want error")
	}

	if len(errs.Errors()) != 2 {
		t.Errorf("Errors() length = %d, want 2", len(errs.Errors()))
	}
}

func TestCustomValidator(t *testing.T) {
	// Register custom validator
	err := RegisterValidator("custom", func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return errors.New("expected string")
		}
		if str != "valid" {
			return errors.New("value must be 'valid'")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RegisterValidator() error = %v", err)
	}
	defer UnregisterValidator("custom")

	type Config struct {
		Field string `validate:"required,custom"`
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{"valid", Config{Field: "valid"}, false},
		{"invalid", Config{Field: "invalid"}, true},
		{"missing", Config{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterValidator_Errors(t *testing.T) {
	tests := []struct {
		name    string
		valName string
		fn      CustomValidatorFunc
		wantErr bool
	}{
		{"empty name", "", func(interface{}) error { return nil }, true},
		{"nil function", "test", nil, true},
		{"builtin conflict", "required", func(interface{}) error { return nil }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterValidator(tt.valName, tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

type testStringer string

func (t testStringer) String() string {
	return string(t)
}
