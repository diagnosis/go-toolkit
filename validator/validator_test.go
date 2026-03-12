package validator

import (
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func assertNoErrors(t *testing.T, v *Validator) {
	t.Helper()
	if err := v.Errors(); err != nil {
		t.Errorf("expected no errors, got: %v", err.Details)
	}
}

func assertError(t *testing.T, v *Validator, field string) {
	t.Helper()
	err := v.Errors()
	if err == nil {
		t.Fatalf("expected error for field %q, got nil", field)
	}
	if _, ok := err.Details[field]; !ok {
		t.Errorf("expected error for field %q, got details: %v", field, err.Details)
	}
}

// ── Required ─────────────────────────────────────────────────────────────────

func TestRequired_Valid(t *testing.T) {
	v := New()
	v.Required("name", "Safa")
	assertNoErrors(t, v)
}

func TestRequired_Empty(t *testing.T) {
	v := New()
	v.Required("name", "")
	assertError(t, v, "name")
}

func TestRequired_Whitespace(t *testing.T) {
	v := New()
	v.Required("name", "   ")
	assertError(t, v, "name")
}

func TestRequired_Tab(t *testing.T) {
	v := New()
	v.Required("name", "\t")
	assertError(t, v, "name")
}

// ── Email ─────────────────────────────────────────────────────────────────────

func TestEmail_Valid(t *testing.T) {
	emails := []string{
		"safa@example.com",
		"safa.d@example.co.uk",
		"safa+tag@gmail.com",
	}
	for _, email := range emails {
		v := New()
		v.Email("email", email)
		assertNoErrors(t, v)
	}
}

func TestEmail_Invalid(t *testing.T) {
	emails := []string{
		"notanemail",
		"missing@",
		"@nodomain.com",
		"spaces in@email.com",
	}
	for _, email := range emails {
		v := New()
		v.Email("email", email)
		assertError(t, v, "email")
	}
}

func TestEmail_NameFormat_Rejected(t *testing.T) {
	// mail.ParseAddress accepts "Name <email>" — we should not
	v := New()
	v.Email("email", "Safa D <safa@example.com>")
	assertError(t, v, "email")
}

func TestEmail_Empty_Skipped(t *testing.T) {
	// empty email is not an email error — use Required for that
	v := New()
	v.Email("email", "")
	assertNoErrors(t, v)
}

// ── Password ──────────────────────────────────────────────────────────────────

func TestPassword_Valid(t *testing.T) {
	v := New()
	v.Password("password", "Secure123")
	assertNoErrors(t, v)
}

func TestPassword_TooShort(t *testing.T) {
	v := New()
	v.Password("password", "Ab1")
	assertError(t, v, "password")
}

func TestPassword_TooLong(t *testing.T) {
	v := New()
	v.Password("password", strings.Repeat("Ab1", 50)) // 150 chars
	assertError(t, v, "password")
}

func TestPassword_NoUppercase(t *testing.T) {
	v := New()
	v.Password("password", "lowercase1")
	assertError(t, v, "password")
}

func TestPassword_NoLowercase(t *testing.T) {
	v := New()
	v.Password("password", "UPPERCASE1")
	assertError(t, v, "password")
}

func TestPassword_NoDigit(t *testing.T) {
	v := New()
	v.Password("password", "NoDigitsHere")
	assertError(t, v, "password")
}

func TestPassword_Empty_Skipped(t *testing.T) {
	v := New()
	v.Password("password", "")
	assertNoErrors(t, v)
}

func TestPassword_Unicode(t *testing.T) {
	// unicode chars should count as runes not bytes
	v := New()
	v.Password("password", "Şifre123") // Turkish chars
	assertNoErrors(t, v)
}

// ── MinLength ─────────────────────────────────────────────────────────────────

func TestMinLength_Valid(t *testing.T) {
	v := New()
	v.MinLength("name", "Safa", 3)
	assertNoErrors(t, v)
}

func TestMinLength_TooShort(t *testing.T) {
	v := New()
	v.MinLength("name", "Sa", 3)
	assertError(t, v, "name")
}

func TestMinLength_Exact(t *testing.T) {
	v := New()
	v.MinLength("name", "Saf", 3)
	assertNoErrors(t, v)
}

func TestMinLength_Empty_Skipped(t *testing.T) {
	v := New()
	v.MinLength("name", "", 3)
	assertNoErrors(t, v)
}

// ── MaxLength ─────────────────────────────────────────────────────────────────

func TestMaxLength_Valid(t *testing.T) {
	v := New()
	v.MaxLength("name", "Safa", 10)
	assertNoErrors(t, v)
}

func TestMaxLength_TooLong(t *testing.T) {
	v := New()
	v.MaxLength("name", "Safa Demirkan", 5)
	assertError(t, v, "name")
}

func TestMaxLength_Exact(t *testing.T) {
	v := New()
	v.MaxLength("name", "Safa", 4)
	assertNoErrors(t, v)
}

// ── Min (numeric) ─────────────────────────────────────────────────────────────

func TestMin_Valid(t *testing.T) {
	v := New()
	v.Min("price", 10.0, 0)
	assertNoErrors(t, v)
}

func TestMin_Zero_Valid(t *testing.T) {
	v := New()
	v.Min("stock", 0, 0) // stock can be 0
	assertNoErrors(t, v)
}

func TestMin_BelowMin(t *testing.T) {
	v := New()
	v.Min("price", -1.0, 0)
	assertError(t, v, "price")
}

// ── multiple fields ───────────────────────────────────────────────────────────

func TestValidator_MultipleFields(t *testing.T) {
	v := New()
	v.Required("name", "")
	v.Email("email", "notanemail")
	v.Password("password", "weak")

	err := v.Errors()
	if err == nil {
		t.Fatal("expected errors, got nil")
	}
	if len(err.Details) != 3 {
		t.Errorf("expected 3 field errors, got %d: %v", len(err.Details), err.Details)
	}
}

func TestValidator_FirstErrorPerField(t *testing.T) {
	// calling the same field twice should keep first error only
	v := New()
	v.Required("email", "")
	v.Email("email", "")

	err := v.Errors()
	if err == nil {
		t.Fatal("expected error")
	}
	if len(err.Details) != 1 {
		t.Errorf("expected 1 error for email field, got %d", len(err.Details))
	}
}

func TestValidator_NoErrors_ReturnsNil(t *testing.T) {
	v := New()
	v.Required("name", "Safa")
	v.Email("email", "safa@example.com")
	v.Password("password", "Secure123")

	if v.Errors() != nil {
		t.Error("expected nil when all fields valid")
	}
}

// ── humanize ─────────────────────────────────────────────────────────────────

func TestHumanize_SingleWord(t *testing.T) {
	if humanize("email") != "Email" {
		t.Errorf("expected Email, got %s", humanize("email"))
	}
}

func TestHumanize_SnakeCase(t *testing.T) {
	if humanize("first_name") != "First Name" {
		t.Errorf("expected First Name, got %s", humanize("first_name"))
	}
}

func TestHumanize_Empty(t *testing.T) {
	if humanize("") != "Field" {
		t.Errorf("expected Field, got %s", humanize(""))
	}
}
