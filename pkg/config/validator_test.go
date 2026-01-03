package config

import (
	"reflect"
	"testing"
)

func TestRequiredValidator(t *testing.T) {
	validator := NewRequiredValidator()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "valid string",
			value:   "test",
			wantErr: false,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only string",
			value:   "   ",
			wantErr: true,
		},
		{
			name:    "valid int",
			value:   int(42),
			wantErr: false,
		},
		{
			name:    "zero int",
			value:   int(0),
			wantErr: true,
		},
		{
			name:    "valid int64",
			value:   int64(100),
			wantErr: false,
		},
		{
			name:    "zero int64",
			value:   int64(0),
			wantErr: true,
		},
		{
			name:    "valid uint",
			value:   uint(42),
			wantErr: false,
		},
		{
			name:    "zero uint",
			value:   uint(0),
			wantErr: true,
		},
		{
			name:    "valid float64",
			value:   float64(3.14),
			wantErr: false,
		},
		{
			name:    "zero float64",
			value:   float64(0),
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			wantErr: true,
		},
		{
			name:    "valid slice",
			value:   []string{"a", "b"},
			wantErr: false,
		},
		{
			name:    "empty slice",
			value:   []string{},
			wantErr: true,
		},
		{
			name:    "valid map",
			value:   map[string]int{"a": 1},
			wantErr: false,
		},
		{
			name:    "empty map",
			value:   map[string]int{},
			wantErr: true,
		},
		{
			name:    "valid pointer",
			value:   &[]int{1, 2},
			wantErr: false,
		},
		{
			name:    "nil pointer",
			value:   (*int)(nil),
			wantErr: true,
		},
		{
			name:    "bool false",
			value:   false,
			wantErr: false, // Bool false is considered valid
		},
		{
			name:    "bool true",
			value:   true,
			wantErr: false,
		},
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
		{
			name:    "valid email",
			value:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			value:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			value:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dot",
			value:   "user.name@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email - no @",
			value:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "invalid email - no domain",
			value:   "user@",
			wantErr: true,
		},
		{
			name:    "invalid email - no local part",
			value:   "@example.com",
			wantErr: true,
		},
		{
			name:    "invalid email - no TLD",
			value:   "user@example",
			wantErr: true,
		},
		{
			name:    "invalid email - spaces",
			value:   "user @example.com",
			wantErr: true,
		},
		{
			name:    "non-string value",
			value:   123,
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
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
		{
			name:    "valid http URL",
			value:   "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			value:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid URL with path",
			value:   "https://example.com/path/to/resource",
			wantErr: false,
		},
		{
			name:    "valid URL with query",
			value:   "https://example.com?key=value",
			wantErr: false,
		},
		{
			name:    "valid URL with port",
			value:   "http://example.com:8080",
			wantErr: false,
		},
		{
			name:    "invalid URL - no scheme",
			value:   "example.com",
			wantErr: true,
		},
		{
			name:    "invalid URL - no host",
			value:   "http://",
			wantErr: true,
		},
		{
			name:    "invalid URL - malformed",
			value:   "://example.com",
			wantErr: true,
		},
		{
			name:    "non-string value",
			value:   123,
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
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
		{
			name:    "valid int in range",
			min:     floatPtr(1),
			max:     floatPtr(100),
			value:   int(50),
			wantErr: false,
		},
		{
			name:    "int at minimum",
			min:     floatPtr(1),
			max:     floatPtr(100),
			value:   int(1),
			wantErr: false,
		},
		{
			name:    "int at maximum",
			min:     floatPtr(1),
			max:     floatPtr(100),
			value:   int(100),
			wantErr: false,
		},
		{
			name:    "int below minimum",
			min:     floatPtr(1),
			max:     floatPtr(100),
			value:   int(0),
			wantErr: true,
		},
		{
			name:    "int above maximum",
			min:     floatPtr(1),
			max:     floatPtr(100),
			value:   int(101),
			wantErr: true,
		},
		{
			name:    "valid float64 in range",
			min:     floatPtr(0.0),
			max:     floatPtr(1.0),
			value:   float64(0.5),
			wantErr: false,
		},
		{
			name:    "float64 below minimum",
			min:     floatPtr(0.0),
			max:     floatPtr(1.0),
			value:   float64(-0.1),
			wantErr: true,
		},
		{
			name:    "only minimum constraint",
			min:     floatPtr(1),
			max:     nil,
			value:   int(50),
			wantErr: false,
		},
		{
			name:    "only minimum constraint - fails",
			min:     floatPtr(1),
			max:     nil,
			value:   int(0),
			wantErr: true,
		},
		{
			name:    "only maximum constraint",
			min:     nil,
			max:     floatPtr(100),
			value:   int(50),
			wantErr: false,
		},
		{
			name:    "only maximum constraint - fails",
			min:     nil,
			max:     floatPtr(100),
			value:   int(101),
			wantErr: true,
		},
		{
			name:    "valid uint",
			min:     floatPtr(1),
			max:     floatPtr(65535),
			value:   uint(8080),
			wantErr: false,
		},
		{
			name:    "non-numeric value",
			min:     floatPtr(1),
			max:     floatPtr(100),
			value:   "not a number",
			wantErr: true,
		},
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

func TestParseTagOptions(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected map[string]string
	}{
		{
			name:     "empty tag",
			tag:      "",
			expected: map[string]string{},
		},
		{
			name:     "single boolean flag",
			tag:      "required",
			expected: map[string]string{"required": ""},
		},
		{
			name:     "single key-value",
			tag:      "env=PORT",
			expected: map[string]string{"env": "PORT"},
		},
		{
			name: "multiple options",
			tag:  "env=PORT,default=8080,required",
			expected: map[string]string{
				"env":      "PORT",
				"default":  "8080",
				"required": "",
			},
		},
		{
			name: "range value with comma",
			tag:  "validate=range=1,65535",
			expected: map[string]string{
				"validate": "range=1,65535",
			},
		},
		{
			name: "multiple options with range",
			tag:  "env=PORT,default=8080,validate=range=1,65535",
			expected: map[string]string{
				"env":      "PORT",
				"default":  "8080",
				"validate": "range=1,65535",
			},
		},
		{
			name: "range with whitespace",
			tag:  "validate=range=1, 65535",
			expected: map[string]string{
				"validate": "range=1,65535", // Whitespace is trimmed
			},
		},
		{
			name: "multiple options with whitespace",
			tag:  "env=PORT , default=8080 , required",
			expected: map[string]string{
				"env":      "PORT",
				"default":  "8080",
				"required": "",
			},
		},
		{
			name: "email validation",
			tag:  "required,validate=email",
			expected: map[string]string{
				"required": "",
				"validate": "email",
			},
		},
		{
			name: "url validation",
			tag:  "validate=url",
			expected: map[string]string{
				"validate": "url",
			},
		},
		{
			name: "range with negative numbers",
			tag:  "validate=range=-100,100",
			expected: map[string]string{
				"validate": "range=-100,100",
			},
		},
		{
			name: "range with decimals",
			tag:  "validate=range=0.0,1.0",
			expected: map[string]string{
				"validate": "range=0.0,1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTagOptions(tt.tag)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseTagOptions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitTagIntoParts(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected []string
	}{
		{
			name:     "empty tag",
			tag:      "",
			expected: []string{""},
		},
		{
			name:     "single part",
			tag:      "required",
			expected: []string{"required"},
		},
		{
			name:     "multiple parts",
			tag:      "env=PORT,default=8080,required",
			expected: []string{"env=PORT", "default=8080", "required"},
		},
		{
			name:     "with whitespace",
			tag:      "env=PORT , default=8080 , required",
			expected: []string{"env=PORT", "default=8080", "required"}, // Whitespace is trimmed
		},
		{
			name:     "range with comma",
			tag:      "validate=range=1,65535",
			expected: []string{"validate=range=1", "65535"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitTagIntoParts(tt.tag)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitTagIntoParts() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMergeRangeValueParts(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected []string
	}{
		{
			name:     "empty parts",
			parts:    []string{},
			expected: []string{},
		},
		{
			name:     "no range values",
			parts:    []string{"env=PORT", "default=8080"},
			expected: []string{"env=PORT", "default=8080"},
		},
		{
			name:     "range value to merge",
			parts:    []string{"validate=range=1", "65535"},
			expected: []string{"validate=range=1,65535"},
		},
		{
			name:     "range value already complete",
			parts:    []string{"validate=range=1,65535"},
			expected: []string{"validate=range=1,65535"},
		},
		{
			name:     "multiple parts with range",
			parts:    []string{"env=PORT", "validate=range=1", "65535", "required"},
			expected: []string{"env=PORT", "validate=range=1,65535", "required"},
		},
		{
			name:     "range with negative number",
			parts:    []string{"validate=range=-100", "100"},
			expected: []string{"validate=range=-100,100"},
		},
		{
			name:     "range with decimal",
			parts:    []string{"validate=range=0.0", "1.0"},
			expected: []string{"validate=range=0.0,1.0"},
		},
		{
			name:     "empty parts filtered",
			parts:    []string{"env=PORT", "", "default=8080"},
			expected: []string{"env=PORT", "default=8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeRangeValueParts(tt.parts)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mergeRangeValueParts() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsRangeValueContinuation(t *testing.T) {
	tests := []struct {
		name     string
		prevPart string
		currPart string
		expected bool
	}{
		{
			name:     "range continuation with number",
			prevPart: "validate=range=1",
			currPart: "65535",
			expected: true,
		},
		{
			name:     "range continuation with negative number",
			prevPart: "validate=range=-100",
			currPart: "100",
			expected: true,
		},
		{
			name:     "range already complete",
			prevPart: "validate=range=1,65535",
			currPart: "something",
			expected: false,
		},
		{
			name:     "not a range value",
			prevPart: "env=PORT",
			currPart: "8080",
			expected: false,
		},
		{
			name:     "range= but not continuation",
			prevPart: "validate=range=",
			currPart: "not-a-number",
			expected: false,
		},
		{
			name:     "range= with whitespace",
			prevPart: "validate=range=1",
			currPart: " 65535",
			expected: true,
		},
		{
			name:     "range= with decimal",
			prevPart: "validate=range=0.0",
			currPart: "1.0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRangeValueContinuation(tt.prevPart, tt.currPart)
			if result != tt.expected {
				t.Errorf("isRangeValueContinuation() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseKeyValuePair(t *testing.T) {
	tests := []struct {
		name        string
		part        string
		expectedKey string
		expectedVal string
	}{
		{
			name:        "empty part",
			part:        "",
			expectedKey: "",
			expectedVal: "",
		},
		{
			name:        "boolean flag",
			part:        "required",
			expectedKey: "required",
			expectedVal: "",
		},
		{
			name:        "key-value pair",
			part:        "env=PORT",
			expectedKey: "env",
			expectedVal: "PORT",
		},
		{
			name:        "key-value with whitespace",
			part:        "env = PORT ",
			expectedKey: "env",
			expectedVal: "PORT",
		},
		{
			name:        "key-value with equals in value",
			part:        "validate=range=1,65535",
			expectedKey: "validate",
			expectedVal: "range=1,65535",
		},
		{
			name:        "only equals sign",
			part:        "=",
			expectedKey: "",
			expectedVal: "",
		},
		{
			name:        "key with no value",
			part:        "key=",
			expectedKey: "key",
			expectedVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val := parseKeyValuePair(tt.part)
			if key != tt.expectedKey || val != tt.expectedVal {
				t.Errorf("parseKeyValuePair() = (%q, %q), want (%q, %q)", key, val, tt.expectedKey, tt.expectedVal)
			}
		})
	}
}

func TestParseRule(t *testing.T) {
	tests := []struct {
		name         string
		rule         string
		expectedName string
		expectedVal  string
	}{
		{
			name:         "email rule",
			rule:         "email",
			expectedName: "email",
			expectedVal:  "",
		},
		{
			name:         "url rule",
			rule:         "url",
			expectedName: "url",
			expectedVal:  "",
		},
		{
			name:         "range rule",
			rule:         "range=1,100",
			expectedName: "range",
			expectedVal:  "1,100",
		},
		{
			name:         "range rule with whitespace",
			rule:         "range = 1,100",
			expectedName: "range",
			expectedVal:  "1,100",
		},
		{
			name:         "empty rule",
			rule:         "",
			expectedName: "",
			expectedVal:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, val := parseRule(tt.rule)
			if name != tt.expectedName || val != tt.expectedVal {
				t.Errorf("parseRule() = (%q, %q), want (%q, %q)", name, val, tt.expectedName, tt.expectedVal)
			}
		})
	}
}

func TestParseRangeValues(t *testing.T) {
	tests := []struct {
		name      string
		ruleValue string
		wantMin   *float64
		wantMax   *float64
		wantErr   bool
	}{
		{
			name:      "min and max",
			ruleValue: "1,100",
			wantMin:   floatPtr(1),
			wantMax:   floatPtr(100),
			wantErr:   false,
		},
		{
			name:      "only min",
			ruleValue: "1,",
			wantMin:   floatPtr(1),
			wantMax:   nil,
			wantErr:   false,
		},
		{
			name:      "only max",
			ruleValue: ",100",
			wantMin:   nil,
			wantMax:   floatPtr(100),
			wantErr:   false,
		},
		{
			name:      "negative numbers",
			ruleValue: "-100,100",
			wantMin:   floatPtr(-100),
			wantMax:   floatPtr(100),
			wantErr:   false,
		},
		{
			name:      "decimals",
			ruleValue: "0.0,1.0",
			wantMin:   floatPtr(0.0),
			wantMax:   floatPtr(1.0),
			wantErr:   false,
		},
		{
			name:      "with whitespace",
			ruleValue: " 1 , 100 ",
			wantMin:   floatPtr(1),
			wantMax:   floatPtr(100),
			wantErr:   false,
		},
		{
			name:      "invalid min",
			ruleValue: "abc,100",
			wantMin:   nil,
			wantMax:   nil,
			wantErr:   true,
		},
		{
			name:      "invalid max",
			ruleValue: "1,abc",
			wantMin:   nil,
			wantMax:   nil,
			wantErr:   true,
		},
		{
			name:      "empty string",
			ruleValue: "",
			wantMin:   nil,
			wantMax:   nil,
			wantErr:   false,
		},
		{
			name:      "only comma",
			ruleValue: ",",
			wantMin:   nil,
			wantMax:   nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max, err := parseRangeValues(tt.ruleValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRangeValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !floatPtrEqual(min, tt.wantMin) {
					t.Errorf("parseRangeValues() min = %v, want %v", min, tt.wantMin)
				}
				if !floatPtrEqual(max, tt.wantMax) {
					t.Errorf("parseRangeValues() max = %v, want %v", max, tt.wantMax)
				}
			}
		})
	}
}

func TestValidateByRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "valid email",
			rule:    "email",
			value:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email",
			rule:    "email",
			value:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "valid URL",
			rule:    "url",
			value:   "https://example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			rule:    "url",
			value:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "valid range",
			rule:    "range=1,100",
			value:   50,
			wantErr: false,
		},
		{
			name:    "invalid range - too low",
			rule:    "range=1,100",
			value:   0,
			wantErr: true,
		},
		{
			name:    "invalid range - too high",
			rule:    "range=1,100",
			value:   101,
			wantErr: true,
		},
		{
			name:    "unknown rule",
			rule:    "unknown",
			value:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateByRule(tt.rule, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateByRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStruct_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "not a struct",
			input:   "not a struct",
			wantErr: true,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "pointer to non-struct",
			input:   &[]int{1, 2},
			wantErr: true,
		},
		{
			name: "struct with no tags",
			input: struct {
				Field string
			}{Field: "value"},
			wantErr: false,
		},
		{
			name: "struct with unexported field",
			input: struct {
				Exported   string `config:"required"`
				unexported string `config:"required"`
			}{
				Exported:   "value",
				unexported: "value",
			},
			wantErr: false, // unexported fields are skipped
		},
		{
			name: "nested struct",
			input: struct {
				Outer string `config:"required"`
				Inner struct {
					InnerField string `config:"required"`
				}
			}{
				Outer: "value",
				Inner: struct {
					InnerField string `config:"required"`
				}{
					InnerField: "value",
				},
			},
			wantErr: false,
		},
		{
			name: "nested struct with validation error",
			input: struct {
				Outer string `config:"required"`
				Inner struct {
					InnerField string `config:"required"`
				}
			}{
				Outer: "value",
				Inner: struct {
					InnerField string `config:"required"`
				}{
					InnerField: "", // empty, should fail
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "Email",
		Value:   "invalid-email",
		Message: "invalid email format",
	}

	expected := `validation error for field "Email": invalid email format (value: invalid-email)`
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}

// Helper functions

func floatPtr(f float64) *float64 {
	return &f
}

func floatPtrEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
