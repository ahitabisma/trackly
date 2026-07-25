package httpx

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func Success(data interface{}, message string) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(),
	}
}

func Error(code ErrorCode, message string, detail string) Response {
	return Response{
		Success: false,
		Message: message,
		Errors: ErrorResponse{
			Code:   code,
			Detail: detail,
		},
		Meta: newMeta(),
	}
}

func ValidationError(fields map[string]string) Response {
	return Response{
		Success: false,
		Message: "Validation Failed",
		Errors: ErrorResponse{
			Code:   ErrValidation,
			Fields: fields,
		},
		Meta: newMeta(),
	}
}

func SuccessWithPagination(
	data interface{},
	message string,
	total, page, limit int,
) Response {

	totalPages := (total + limit - 1) / limit

	return Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &Meta{
			RequestID: uuid.NewString(),
			Timestamp: time.Now(),
			Pagination: &Pagination{
				Total:       total,
				Page:        page,
				Limit:       limit,
				TotalPages:  totalPages,
				HasNextPage: page < totalPages,
				HasPrevPage: page > 1,
			},
		},
	}
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, res Response) {
	// res.Meta = newMetaFromCtx(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}