package filter

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func ApplyGormFilter(db *gorm.DB, fq FilteringQuery, allowedFields []string) *gorm.DB {
	allowed := make(map[string]bool, len(allowedFields))
	for _, f := range allowedFields {
		allowed[f] = true
	}

	// ── exact filters
	for key, val := range fq.Filters {
		if !allowed[key] || val == nil {
			continue
		}
		db = db.Where(fmt.Sprintf("%s = ?", key), val)
	}

	// ── search filters (ILIKE)
	if fq.SearchFilters != nil {
		switch sf := fq.SearchFilters.(type) {

		case []SearchFilterItem:
			var conditions []string
			var args []interface{}

			for _, item := range sf {
				if !allowed[item.Field] || item.Value == "" {
					continue
				}
				conditions = append(conditions, fmt.Sprintf("%s ILIKE ?", item.Field))
				args = append(args, "%"+fmt.Sprintf("%v", item.Value)+"%")
			}

			if len(conditions) > 0 {
				db = db.Where("("+strings.Join(conditions, " OR ")+")", args...)
			}

		case map[string]interface{}:
			for key, val := range sf {
				if !allowed[key] || val == "" {
					continue
				}
				db = db.Where(fmt.Sprintf("%s ILIKE ?", key), "%"+fmt.Sprintf("%v", val)+"%")
			}
		}
	}

	// ── ranged filters
	for _, rf := range fq.RangedFilters {
		if !allowed[rf.Key] {
			continue
		}

		if rf.From != nil && rf.To != nil {
			db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", rf.Key), rf.From, rf.To)
		} else if rf.From != nil {
			db = db.Where(fmt.Sprintf("%s >= ?", rf.Key), rf.From)
		} else if rf.To != nil {
			db = db.Where(fmt.Sprintf("%s <= ?", rf.Key), rf.To)
		}
	}

	return db
}

func ApplyPagination(db *gorm.DB, fq FilteringQuery) (*gorm.DB, int, int) {
	limit := fq.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	page := fq.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	return db.Limit(limit).Offset(offset), page, limit
}

func ParseFromRequest(r *http.Request) FilteringQuery {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	page, _ := strconv.Atoi(q.Get("page"))

	// Set default and max limit
	if limit <= 0 {
		limit = 10
	}
	limit = int(math.Min(float64(limit), 100)) // Max 100 per page

	// Set default page
	if page <= 0 {
		page = 1
	}

	fq := FilteringQuery{
		Page:      page,
		Limit:     limit,
		OrderKey:  q.Get("orderKey"),
		OrderRule: q.Get("orderRule"),
	}

	// filters: {"status":"active","role":"admin"}
	if v := q.Get("filters"); v != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(v), &m); err == nil {
			fq.Filters = m
		}
	}

	// searchFilters: array of {"field":"name","value":"text"} or object {"name":"text"}
	if v := q.Get("searchFilters"); v != "" {
		var arr []SearchFilterItem
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			fq.SearchFilters = arr
		} else {
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(v), &obj); err == nil {
				fq.SearchFilters = obj
			}
		}
	}

	// rangedFilters: [{"key":"created_at","from":"2024-01-01","to":"2024-12-31"}]
	if v := q.Get("rangedFilters"); v != "" {
		var ranges []RangedFilter
		if err := json.Unmarshal([]byte(v), &ranges); err == nil {
			fq.RangedFilters = ranges
		}
	}

	return fq
}

func WrapPaginated[T any](data []T, total int64, page, limit int) PaginatedResult[T] {
	if limit <= 0 {
		limit = 10
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	hasNextPage := page < totalPages
	hasPrevPage := page > 1

	return PaginatedResult[T]{
		Data: data,
		Pagination: Pagination{
			Total:       total,
			Page:        page,
			Limit:       limit,
			TotalPages:  totalPages,
			HasNextPage: hasNextPage,
			HasPrevPage: hasPrevPage,
		},
	}
}
