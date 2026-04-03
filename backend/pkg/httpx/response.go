package httpx

import (
	"encoding/json"
	"net/http"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorContent struct {
	Code        string       `json:"code"`
	FieldErrors []FieldError `json:"field_errors,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Content interface{} `json:"content"`
}

func SuccessWithoutPagination(w http.ResponseWriter, data interface{}, message string) {
	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Content: data,
	})
}

func Created(w http.ResponseWriter, data interface{}, message string) {
	sendJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: message,
		Content: data,
	})
}

func Error(w http.ResponseWriter, statusCode int, errorCode string, message string) {
	content := ErrorContent{
		Code:        errorCode,
		FieldErrors: []FieldError{},
	}

	sendJSON(w, statusCode, Response{
		Success: false,
		Message: message,
		Content: content,
	})
}

func ValidationError(w http.ResponseWriter, message string, fieldErrors []FieldError) {
	if fieldErrors == nil {
		fieldErrors = []FieldError{}
	}

	content := ErrorContent{
		Code:        "VALIDATION_ERROR",
		FieldErrors: fieldErrors,
	}

	sendJSON(w, http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Content: content,
	})
}

func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
