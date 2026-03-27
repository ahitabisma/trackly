package filter

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Options menyimpan konfigurasi filter, sort, dan paginasi dari URL query.
type Options struct {
	OrderKey      string                 `json:"orderKey"`
	OrderRule     string                 `json:"orderRule"`
	Filters       map[string]interface{} `json:"filters"`
	SearchFilters []SearchFilterItem     `json:"searchFilters"`
	RangedFilters []RangedFilter         `json:"rangedFilters"`
	Limit         int                    `json:"limit"`
	Page          int                    `json:"page"`
	Q             string                 `json:"q"`
	GroupBy       string                 `json:"groupBy"`
}

type RangedFilter struct {
	Key  string      `json:"key"`
	From interface{} `json:"from"`
	To   interface{} `json:"to"`
}

type SearchFilterItem struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

type PaginatedResult[T any] struct {
	Data       []T           `json:"data"`
	Pagination PaginationRes `json:"pagination"`
}

type PaginationRes struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalPages  int  `json:"totalPages"`
	HasNextPage bool `json:"hasNextPage"`
	HasPrevPage bool `json:"hasPrevPage"`
}

// ParseQuery membaca parameter http.Request URL dan mengonversinya menjadi Options.
func ParseQuery(r *http.Request) Options {
	q := r.URL.Query()
	opt := Options{}

	if v := q.Get("orderKey"); v != "" {
		opt.OrderKey = v
	}
	if v := q.Get("orderRule"); v != "" {
		opt.OrderRule = v
	}
	if v := q.Get("filters"); v != "" {
		_ = json.Unmarshal([]byte(v), &opt.Filters)
	}
	if v := q.Get("searchFilters"); v != "" {
		_ = json.Unmarshal([]byte(v), &opt.SearchFilters)
	}
	if v := q.Get("rangedFilters"); v != "" {
		_ = json.Unmarshal([]byte(v), &opt.RangedFilters)
	}
	if v := q.Get("limit"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil {
			opt.Limit = limit
		}
	}
	if v := q.Get("page"); v != "" {
		if page, err := strconv.Atoi(v); err == nil {
			opt.Page = page
		}
	}
	if v := q.Get("q"); v != "" {
		opt.Q = v
	}
	if v := q.Get("groupBy"); v != "" {
		opt.GroupBy = v
	}
	return opt
}

// BuildMongoQuery mengonversi Options menjadi filter bson.M dan opsi pencari (*options.FindOptionsBuilder) untuk MongoDB.
func BuildMongoQuery(opt Options, allowedFields []string) (bson.M, *options.FindOptionsBuilder) {
	mongoFilter := bson.M{}
	findOpts := options.Find()

	// 1. FILTERING (Exact Match atau IN)
	if opt.Filters != nil {
		for key, value := range opt.Filters {
			if !contains(allowedFields, key) || value == nil {
				continue
			}
			if arr, ok := value.([]interface{}); ok {
				mongoFilter[key] = bson.M{"$in": arr}
			} else {
				mongoFilter[key] = value
			}
		}
	}

	// 2. SEARCH (Regex / Pencarian Teks)
	if len(opt.SearchFilters) > 0 {
		var orConds []bson.M
		for _, item := range opt.SearchFilters {
			if contains(allowedFields, item.Field) {
				orConds = append(orConds, bson.M{
					item.Field: bson.M{"$regex": item.Value, "$options": "i"},
				})
			}
		}
		if len(orConds) > 0 {
			mongoFilter["$or"] = orConds
		}
	}

	// 3. SORTING
	orderKey := opt.OrderKey
	orderRule := opt.OrderRule

	if orderKey == "" || !contains(allowedFields, orderKey) {
		orderKey = "time"  // Default sort based on time
		orderRule = "desc" // Default sort order
	}

	order := -1 // default desc
	if orderRule == "asc" {
		order = 1
	}
	findOpts.SetSort(bson.D{{Key: orderKey, Value: order}})

	// 4. PAGINATION
	limit := int64(10) // Default limit 10 item per halaman
	if opt.Limit > 0 && opt.Limit <= 100 {
		limit = int64(opt.Limit)
	}

	page := int64(1) // Default page 1
	if opt.Page > 0 {
		page = int64(opt.Page)
	}

	skip := (page - 1) * limit
	findOpts.SetLimit(limit)
	findOpts.SetSkip(skip)

	return mongoFilter, findOpts
}

// contains adalah helper internal untuk memeriksa kecocokan target di dalam sebuah array slice string.
// Berguna untuk memastikan bahwa fields yang difilter dan di-sort adalah valid.
func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

// WrapPaginated membungkus respons slice menjadi format paginasi JSON berserta properti informatif.
func WrapPaginated[T any](data []T, total, page, limit int) PaginatedResult[T] {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	return PaginatedResult[T]{
		Data: data,
		Pagination: PaginationRes{
			Total:       total,
			Page:        page,
			Limit:       limit,
			TotalPages:  totalPages,
			HasNextPage: page < totalPages,
			HasPrevPage: page > 1,
		},
	}
}
