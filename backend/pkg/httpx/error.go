package httpx

type ErrorCode string

const (
	ErrInternal           ErrorCode = "INTERNAL_ERROR"
	ErrNotFound           ErrorCode = "NOT_FOUND"
	ErrInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrValidation         ErrorCode = "VALIDATION_ERROR"
	ErrForbidden          ErrorCode = "FORBIDDEN"
	ErrConflict           ErrorCode = "CONFLICT"
	ErrUnauthorized       ErrorCode = "UNAUTHORIZED"
)

type ErrorResponse struct {
	Code   ErrorCode         `json:"code"`
	Detail string            `json:"detail,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}
