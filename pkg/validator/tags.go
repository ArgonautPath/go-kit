package validator

import (
	"strings"
)

const (
	// DefaultTagName is the default struct tag name for validation.
	DefaultTagName = "validate"
)

// parseTagOptions parses struct tag options.
// Handles commas inside values (e.g., "range=1,65535").
// Returns a map of option keys to values.
func parseTagOptions(tag string) map[string]string {
	options := make(map[string]string)

	if tag == "" {
		return options
	}

	// Split by comma, but be careful with commas inside values
	parts := splitTagIntoParts(tag)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, value := parseKeyValuePair(part)
		if key != "" {
			options[key] = value
		}
	}

	return options
}

// splitTagIntoParts splits a tag string into parts, handling commas inside values.
// Example: "required,range=1,65535,email" -> ["required", "range=1,65535", "email"]
func splitTagIntoParts(tag string) []string {
	var parts []string
	var current strings.Builder
	var inValue bool

	for i, char := range tag {
		switch char {
		case '=':
			inValue = true
			current.WriteRune(char)
		case ',':
			if inValue {
				// Check if this comma is part of a value (e.g., range=1,65535 or oneof=a,b,c)
				// Look ahead to see if we're still in a value
				remaining := tag[i+1:]
				if isRangeValueContinuation(current.String(), remaining) || isOneOfValueContinuation(current.String(), remaining) {
					current.WriteRune(char)
					continue
				}
			}
			// End of current part
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
				inValue = false
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// isRangeValueContinuation checks if the remaining string is a continuation of a
// range value yet (missing the comma-separated min,max values).
func isRangeValueContinuation(prevPart, remaining string) bool {
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
	trimmed := strings.TrimSpace(remaining)
	if len(trimmed) > 0 {
		firstChar := trimmed[0]
		if (firstChar >= '0' && firstChar <= '9') || firstChar == '-' {
			return true
		}
	}

	return false
}

// isOneOfValueContinuation checks if the remaining string is a continuation of a
// oneOf value (comma-separated list of allowed values).
func isOneOfValueContinuation(prevPart, remaining string) bool {
	// Previous part must contain "oneof="
	if !strings.Contains(prevPart, "oneof=") {
		return false
	}

	// If we're in a oneOf value, continue until we hit a rule name (starts with letter)
	// This is a heuristic: if remaining doesn't start with a known rule name, it's likely part of oneOf
	trimmed := strings.TrimSpace(remaining)
	if len(trimmed) == 0 {
		return false
	}

	// Known rule names that would indicate end of oneOf value
	knownRules := []string{"required", "email", "url", "range", "len", "min", "max", "regex", "alpha", "alphanumeric", "numeric"}
	for _, rule := range knownRules {
		if strings.HasPrefix(trimmed, rule+"=") || trimmed == rule {
			return false
		}
	}

	// If it doesn't look like a rule, it's probably part of oneOf value
	return true
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
