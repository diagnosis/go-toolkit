// Package apperr defines StatusErr, an application error type that carries an
// application status code, a client-facing message, an internal message, and
// an HTTP status, plus constructors for the common error categories.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// Status is an application-level error code, distinct from the HTTP status,
// that clients can switch on to identify the error category.
type Status int

// Application status codes carried by StatusErr and reported in the JSON
// error envelope.
const (
	CodeBadRequest Status = iota + 1
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeConflict
	CodeInternalError
	CodeDatabaseError
	CodeValidationError
	CodeTokenErr
	CodeInvalidCredentials
	CodeAccountInactive
	CodeEmailExists
	CodeDefaultError
	CodeTooManyRequests
	CodeUnprocessableContent
	CodeEmailNotVerified
)

var codeNames = map[Status]string{
	CodeBadRequest: "bad_request",
	CodeUnauthorized: "unauthorized",
	CodeForbidden: "forbidden",
	CodeNotFound: "not_found",
	CodeConflict: "conflict",
	CodeInternalError: "internal_error",
	CodeDatabaseError: "database_error",
	CodeValidationError: "validation_error",
	CodeTokenErr: "token_error",
	CodeInvalidCredentials: "invalid_credentials",
	CodeAccountInactive: "account_inactive",
	CodeEmailExists: "email_exists",
	CodeDefaultError: "default_error",
	CodeTooManyRequests: "too_many_requests",
	CodeUnprocessableContent: "unprocessable_content",
	CodeEmailNotVerified: "email_not_verified",
}
// Code returns name of error code
func (s Status) Code() string{
	if name, ok := codeNames[s]; ok {
		return name
	}
	return "unknown"
}

// StatusErr is an error enriched with an application Status, a client-facing
// Message, an operator-facing InternalMessage, the HTTP status to respond
// with, an optional wrapped cause, and optional per-field Details.
type StatusErr struct {
	Status          Status
	Code string
	Message         string
	InternalMessage string
	HTTPStatus      int
	Err             error
	Details         map[string]string
}

// Error implements the error interface, combining the HTTP status, internal
// message, client message, and wrapped cause when present.
func (se *StatusErr) Error() string {
	if se.Err != nil {
		return fmt.Sprintf("[%d] %s - %s: %v", se.HTTPStatus, se.InternalMessage, se.Message, se.Err)
	}
	if se.InternalMessage != "" {
		return fmt.Sprintf("[%d] %s - %s", se.HTTPStatus, se.InternalMessage, se.Message)
	}
	return fmt.Sprintf("[%d] %s", se.HTTPStatus, se.Message)
}

// Unwrap returns the wrapped cause, enabling errors.Is and errors.As.
func (se *StatusErr) Unwrap() error {
	return se.Err
}

// New builds a StatusErr from its parts. Prefer the category constructors
// (NotFound, Unauthorized, ...) which choose the Status and HTTP status.
func New(status Status, message, internalMsg string, httpStatus int, err error) *StatusErr {
	return &StatusErr{
		Status:          status,
		Code:  			status.Code(),
		Message:         message,
		InternalMessage: internalMsg,
		HTTPStatus:      httpStatus,
		Err:             err,
	}
}

func unwrapErr(err ...error) error {
	if len(err) > 0 {
		return err[0]
	}
	return nil
}

// BadRequest returns a StatusErr with CodeBadRequest and HTTP 400.
func BadRequest(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeBadRequest, message, internalMsg, http.StatusBadRequest, unwrapErr(err...))
}

// NotFound returns a StatusErr with CodeNotFound and HTTP 404.
func NotFound(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeNotFound, message, internalMsg, http.StatusNotFound, unwrapErr(err...))
}

// Internal returns a StatusErr with CodeInternalError and HTTP 500.
func Internal(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeInternalError, message, internalMsg, http.StatusInternalServerError, unwrapErr(err...))
}

// Unauthorized returns a StatusErr with CodeUnauthorized and HTTP 401.
func Unauthorized(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeUnauthorized, message, internalMsg, http.StatusUnauthorized, unwrapErr(err...))
}

// Forbidden returns a StatusErr with CodeForbidden and HTTP 403.
func Forbidden(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeForbidden, message, internalMsg, http.StatusForbidden, unwrapErr(err...))
}

// Conflict returns a StatusErr with CodeConflict and HTTP 409.
func Conflict(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeConflict, message, internalMsg, http.StatusConflict, unwrapErr(err...))
}

// Validation returns a StatusErr with CodeValidationError and HTTP 400.
func Validation(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeValidationError, message, internalMsg, http.StatusBadRequest, unwrapErr(err...))
}

// Database returns a StatusErr with CodeDatabaseError and HTTP 500.
func Database(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeDatabaseError, message, internalMsg, http.StatusInternalServerError, unwrapErr(err...))
}

// TokenError returns a StatusErr with CodeTokenErr and HTTP 401.
func TokenError(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeTokenErr, message, internalMsg, http.StatusUnauthorized, unwrapErr(err...))
}

// InvalidCredentials returns a StatusErr with CodeInvalidCredentials and HTTP 401.
func InvalidCredentials(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeInvalidCredentials, message, internalMsg, http.StatusUnauthorized, unwrapErr(err...))
}

// AccountInactive returns a StatusErr with CodeAccountInactive and HTTP 403.
func AccountInactive(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeAccountInactive, message, internalMsg, http.StatusForbidden, unwrapErr(err...))
}

// EmailExists returns a StatusErr with CodeEmailExists and HTTP 409.
func EmailExists(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeEmailExists, message, internalMsg, http.StatusConflict, unwrapErr(err...))
}

// DefaultError returns a StatusErr with CodeDefaultError and HTTP 500,
// used as the fallback for errors that are not a StatusErr.
func DefaultError(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeDefaultError, message, internalMsg, http.StatusInternalServerError, unwrapErr(err...))
}

// TooManyRequests returns a StatusErr with CodeTooManyRequests and HTTP 429.
func TooManyRequests(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeTooManyRequests, message, internalMsg, http.StatusTooManyRequests, unwrapErr(err...))
}

// ValidationDetails returns a validation StatusErr (CodeValidationError,
// HTTP 400) carrying per-field error messages in Details.
func ValidationDetails(message, internalMsg string, details map[string]string) *StatusErr {
	se := New(CodeValidationError, message, internalMsg, http.StatusBadRequest, nil)
	se.Details = details
	return se
}
// UnprocessableContent returns a StatusErr with CodeUnprocessableContent and HTTP 422
func UnprocessableContent(message, internalMsg string, err ...error) *StatusErr{
	return New(CodeUnprocessableContent, message, internalMsg, http.StatusUnprocessableEntity, unwrapErr(err...))
}

func EmailNotVerified(message, internalMsg string, err ...error) *StatusErr{
	return New(CodeEmailNotVerified, message, internalMsg, http.StatusForbidden, unwrapErr(err...))
}

// IsStatusErr reports whether err is, or wraps, a *StatusErr.
func IsStatusErr(err error) bool {
	if err == nil {
		return false
	}
	var se *StatusErr
	return errors.As(err, &se)
}

// AsStatusErr extracts a *StatusErr from err (directly or wrapped) and
// reports whether it found one.
func AsStatusErr(err error) (*StatusErr, bool) {
	if err == nil {
		return nil, false
	}
	var se *StatusErr
	if errors.As(err, &se) {
		return se, true
	}
	return nil, false
}
