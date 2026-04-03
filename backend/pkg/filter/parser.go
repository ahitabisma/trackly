package filter

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

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
