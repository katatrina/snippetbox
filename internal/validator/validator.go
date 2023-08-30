package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validator contains a map of validation errors for our
// form fields.
type Validator struct {
	FieldErrors  map[string]string
	GenericError string
}

// EmailRX uses the regexp.MustCompile() function to parse a regular expression pattern
// for sanity checking the format of an email address. This returns a pointer to
// a 'compiled' regexp.Regexp type, or panics in the event of an error. Parsing
// this pattern once at startup and storing the compiled *regexp.Regexp in a
// variable is more performant than re-parsing the pattern each time we need it.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// IsNoErrors returns true if the FieldErrors map doesn't contain any entries.
func (v *Validator) IsNoErrors() bool {
	return len(v.FieldErrors) == 0 && v.GenericError == ""
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

func (v *Validator) AddGenericError(message string) {
	v.GenericError = message
}

func IsNotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// IsStringNotExceedLimit returns true if a value contains no more than n bytes.
func IsStringNotExceedLimit(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// IsStringNotLessThanLimit returns true if a value contains at least n bytes.
func IsStringNotLessThanLimit(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
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

// IsMatchRegex returns true if a value matches a provided compiled regular
// expression pattern.
func IsMatchRegex(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
