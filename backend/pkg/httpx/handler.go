package httpx

import (
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Wrap(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			// Handle Validation Error
			if vErr, ok := err.(*ValidationErrorWrapper); ok {
				ValidationError(w, vErr.Message, vErr.FieldErrors)
				return
			}

			// Handle Custom Error (Unauthorized, Not Found, etc)
			if cErr, ok := err.(*CustomError); ok {
				Error(w, cErr.StatusCode, cErr.Code, cErr.Message)
				return
			}

			// Fallback Internal Error
			InternalError(w, "An unexpected error occurred")
		}
	}
}

type CustomError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(statusCode int, code string, message string) *CustomError {
	return &CustomError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

type ValidationErrorWrapper struct {
	Message     string
	FieldErrors []FieldError
}

func (e *ValidationErrorWrapper) Error() string {
	return e.Message
}

func NewValidationError(message string, fieldErrors []FieldError) *ValidationErrorWrapper {
	return &ValidationErrorWrapper{
		Message:     message,
		FieldErrors: fieldErrors,
	}
}

func NewFieldError(field string, message string) FieldError {
	return FieldError{
		Field:   field,
		Message: message,
	}
}

func NewNotFoundError(message string) *CustomError {
	return NewCustomError(http.StatusNotFound, "NOT_FOUND", message)
}

func NewUnauthorizedError(message string) *CustomError {
	return NewCustomError(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func NewForbiddenError(message string) *CustomError {
	return NewCustomError(http.StatusForbidden, "FORBIDDEN", message)
}

func NewBadRequestError(message string) *CustomError {
	return NewCustomError(http.StatusBadRequest, "BAD_REQUEST", message)
}

func NewConflictError(message string) *CustomError {
	return NewCustomError(http.StatusConflict, "CONFLICT", message)
}
