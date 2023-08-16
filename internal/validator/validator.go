package validator

import (
	"strings"
	"unicode/utf8"
)

// Validator contains a map of validation errors for our
// form fields.
type Validator struct {
	FieldErrors map[string]string
}

// IsNoErrors returns true if the FieldErrors map doesn't contain any entries.
func (v *Validator) IsNoErrors() bool {
	return len(v.FieldErrors) == 0
}

// AddFieldError adds an error message to the FieldErrors map
// (so long as no entry already exists for the given key).
func (v *Validator) AddFieldError(key, message string) {
	// Note: We need to initialize the map first, if it isn't already
	// initialized. Because we can't add key:value to a nil map.
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func IsNotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// IsStringNotExceedLimit returns true if a value contains no more than n bytes.
func IsStringNotExceedLimit(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// IsIntInList returns true if a value is in a list of permitted integers.
func IsIntInList(value int, list ...int) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}

	return false
}
