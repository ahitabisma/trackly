package httpx

import (
	"encoding/json"
	"net/http"
)

// Pagination struct sesuai permintaan
type Pagination struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorPayload struct {
	Code        string       `json:"code"`
	FieldErrors []FieldError `json:"field_errors,omitempty"`
}

// Response adalah struktur tunggal untuk semua API response
type Response struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Code    string        `json:"code,omitempty"`
	Data    interface{}   `json:"data,omitempty"`
	Meta    *Pagination   `json:"meta,omitempty"`
	Errors  *ErrorPayload `json:"errors,omitempty"`
}

// SuccessWithoutPagination untuk respons 200 standar
func Success(w http.ResponseWriter, data interface{}, message string) {
	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithPagination untuk respons list data
func SuccessWithPagination(w http.ResponseWriter, data interface{}, meta Pagination, message string) {
	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func Created(w http.ResponseWriter, data interface{}, message string) {
	sendJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, statusCode int, errorCode string, message string) {
	sendJSON(w, statusCode, Response{
		Success: false,
		Message: message,
		Code:    errorCode, // Opsi 1: Code di level utama
		Errors: &ErrorPayload{
			Code: errorCode, // Opsi 2: Juga ada di dalam errors untuk detail
		},
	})
}

func ValidationError(w http.ResponseWriter, message string, fieldErrors []FieldError) {
	sendJSON(w, http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Code:    "VALIDATION_ERROR",
		Errors: &ErrorPayload{
			Code:        "VALIDATION_ERROR",
			FieldErrors: fieldErrors,
		},
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