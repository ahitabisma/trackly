package filter

type FilteringQuery struct {
	Page          int
	Limit         int
	OrderKey      string
	OrderRule     string
	Filters       map[string]interface{}
	SearchFilters interface{} // []SearchFilterItem or map[string]interface{}
	RangedFilters []RangedFilter
}

// SearchFilterItem represents a single search filter for LIKE queries
type SearchFilterItem struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

// RangedFilter represents a ranged filter for BETWEEN queries
type RangedFilter struct {
	Key  string      `json:"key"`
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
}

// Pagination contains pagination information
type Pagination struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

// PaginatedResult wraps paginated data with pagination metadata
type PaginatedResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}
