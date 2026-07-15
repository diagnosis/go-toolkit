// Package validator accumulates field validation errors (required,
// email, password strength, length, and minimum-value checks) and
// converts them into an apperr validation error.
package validator

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/diagnosis/go-toolkit/v2/apperr"
)

// Validator collects validation errors per field. The zero value is not
// ready to use; construct one with New. Only the first error recorded for
// a field is kept.
type Validator struct {
	errors map[string]string
}

// New returns an empty Validator ready to collect field errors.
func New() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

// Required records an error for field when value is empty or whitespace.
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.addError(field, fmt.Sprintf("%s is required", humanize(field)))
	}
}

// Email records an error for field when value is not a valid email
// address. Empty values are skipped; combine with Required if the field
// is mandatory.
func (v *Validator) Email(field, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	addr, err := mail.ParseAddress(value)
	if err != nil || addr.Address != value {
		v.addError(field, fmt.Sprintf("%s must be a valid email address", humanize(field)))
	}
}

// Password records an error for field when value is shorter than 8 or
// longer than 128 characters, or lacks an uppercase letter, a lowercase
// letter, or a digit. Empty values are skipped.
func (v *Validator) Password(field, value string) {
	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) < 8 {
		v.addError(field, fmt.Sprintf("%s must be at least 8 characters", humanize(field)))
		return
	}
	if utf8.RuneCountInString(value) > 128 {
		v.addError(field, fmt.Sprintf("%s must be at most 128 characters", humanize(field)))
		return
	}

	var hasUpper, hasLower, hasDigit bool

	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		v.addError(
			field,
			fmt.Sprintf("%s must contain at least one uppercase letter, one lowercase letter, and one number", humanize(field)),
		)
	}
}

// MinLength records an error for field when value has fewer than min
// characters. Empty values are skipped.
func (v *Validator) MinLength(field, value string, min int) {
	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) < min {
		v.addError(field, fmt.Sprintf("%s must be at least %d characters", humanize(field), min))
	}
}

// MaxLength records an error for field when value has more than max
// characters. Empty values are skipped.
func (v *Validator) MaxLength(field, value string, max int) {
	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) > max {
		v.addError(field, fmt.Sprintf("%s must be at most %d characters", humanize(field), max))
	}
}

// Min records an error for field when value is less than min.
func (v *Validator) Min(field string, value, min float64) {
	if value < min {
		v.addError(field, fmt.Sprintf("%s must be at least %v", humanize(field), min))
	}
}

// Errors returns the collected errors as an apperr validation error with
// per-field details, or nil when every check passed.
func (v *Validator) Errors() *apperr.StatusErr {
	if len(v.errors) == 0 {
		return nil
	}

	return apperr.ValidationDetails(
		"validation failed",
		"validation failed",
		v.errors,
	)
}

func (v *Validator) addError(field, message string) {
	if v.errors == nil {
		v.errors = make(map[string]string)
	}

	// keep first error per field
	if _, exists := v.errors[field]; exists {
		return
	}

	v.errors[field] = message
}

func humanize(field string) string {
	field = strings.TrimSpace(field)
	if field == "" {
		return "Field"
	}

	parts := strings.Split(field, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}

		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	return strings.Join(parts, " ")
}
