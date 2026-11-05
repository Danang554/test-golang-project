package apperror

import (
	"fmt"
	"net/http"
)

// AppError represents a structured, standardized application-level error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	RawErr  error  `json:"-"`
	Details any    `json:"details,omitempty"`
}

// Error implements the standard Go error interface
func (e *AppError) Error() string {
	if e.RawErr != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.RawErr)
	}
	return e.Message
}

// Unwrap allows errors.Is and errors.As to inspect the underlying error
func (e *AppError) Unwrap() error {
	return e.RawErr
}

// Is implements custom error matching for errors.Is
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Message == t.Message
}

// New creates a custom AppError with an explicit HTTP status code
func New(code int, message string, rawErr error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		RawErr:  rawErr,
	}
}

// BadRequest (400 Bad Request)
func BadRequest(message string, details ...any) *AppError {
	var d any
	if len(details) > 0 {
		d = details[0]
	}
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Details: d,
	}
}

// Unauthorized (401 Unauthorized)
func Unauthorized(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// Forbidden (403 Forbidden)
func Forbidden(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// NotFound (404 Not Found)
func NotFound(message string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

// Conflict (409 Conflict)
func Conflict(message string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: message,
	}
}

// UnprocessableEntity (422 Unprocessable Entity)
func UnprocessableEntity(message string, details ...any) *AppError {
	var d any
	if len(details) > 0 {
		d = details[0]
	}
	return &AppError{
		Code:    http.StatusUnprocessableEntity,
		Message: message,
		Details: d,
	}
}

// Internal (500 Internal Server Error) wraps unexpected errors
func Internal(rawErr error, message ...string) *AppError {
	msg := "Internal server error"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: msg,
		RawErr:  rawErr,
	}
}
