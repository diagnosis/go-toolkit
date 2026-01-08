package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Status int

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
)

type StatusErr struct {
	Status          Status
	Message         string
	InternalMessage string
	HTTPStatus      int
	Err             error
}

func (se *StatusErr) Error() string {
	if se.Err != nil {
		return fmt.Sprintf("[%d] %s - %s: %v", se.HTTPStatus, se.InternalMessage, se.Message, se.Err)
	}
	if se.InternalMessage != "" {
		return fmt.Sprintf("[%d] %s - %s", se.HTTPStatus, se.InternalMessage, se.Message)
	}
	return fmt.Sprintf("[%d] %s", se.HTTPStatus, se.Message)
}
func (se *StatusErr) Unwrap() error {
	return se.Err
}
func New(status Status, message, internalMsg string, httpStatus int, err error) *StatusErr {

	return &StatusErr{
		Status:          status,
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
func BadRequest(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeBadRequest, message, internalMsg, http.StatusBadRequest, unwrapErr(err...))
}

func NotFound(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeNotFound, message, internalMsg, http.StatusNotFound, unwrapErr(err...))
}
func Internal(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeInternalError, message, internalMsg, http.StatusInternalServerError, unwrapErr(err...))
}
func Unauthorized(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeUnauthorized, message, internalMsg, http.StatusUnauthorized, unwrapErr(err...))
}

func Forbidden(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeForbidden, message, internalMsg, http.StatusForbidden, unwrapErr(err...))
}

func Conflict(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeConflict, message, internalMsg, http.StatusConflict, unwrapErr(err...))
}

func Validation(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeValidationError, message, internalMsg, http.StatusBadRequest, unwrapErr(err...))
}

func Database(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeDatabaseError, message, internalMsg, http.StatusInternalServerError, unwrapErr(err...))
}

func TokenError(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeTokenErr, message, internalMsg, http.StatusUnauthorized, unwrapErr(err...))
}

func InvalidCredentials(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeInvalidCredentials, message, internalMsg, http.StatusUnauthorized, unwrapErr(err...))
}

func AccountInactive(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeAccountInactive, message, internalMsg, http.StatusForbidden, unwrapErr(err...))
}

func EmailExists(message, internalMsg string, err ...error) *StatusErr {

	return New(CodeEmailExists, message, internalMsg, http.StatusConflict, unwrapErr(err...))
}

func DefaultError(message, internalMsg string, err ...error) *StatusErr {
	return New(CodeDefaultError, message, internalMsg, http.StatusInternalServerError, unwrapErr(err...))
}

func IsStatusErr(err error) bool {
	if err == nil {
		return false
	}
	var se *StatusErr
	return errors.As(err, &se)
}
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
