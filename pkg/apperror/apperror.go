package apperror

import (
	"errors"
	"net/http"
)

type AppError struct {
	Status  int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, msg string) *AppError {
	return &AppError{Status: status, Message: msg}
}

func Wrap(err error, status int, msg string) *AppError {
	return &AppError{Status: status, Message: msg, Err: err}
}

func BadRequest(msg string) *AppError { return New(http.StatusBadRequest, msg) }
func NotFound(msg string) *AppError   { return New(http.StatusNotFound, msg) }
func Conflict(msg string) *AppError   { return New(http.StatusConflict, msg) }
func Internal(msg string) *AppError   { return New(http.StatusInternalServerError, msg) }

func Status(err error) int {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Status
	}
	return http.StatusInternalServerError
}

func Message(err error) string {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Message
	}
	return "internal server error"
}
