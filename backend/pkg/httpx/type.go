package httpx

import "time"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type Meta struct {
	RequestID  string      `json:"request_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}
