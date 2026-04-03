package filter

import (
	"fmt"
	"math"
	"strings"
)

type SQLResult struct {
	Where  string
	Args   []interface{}
	Order  string
	Limit  int
	Offset int
}

func BuildSQL(fq FilteringQuery, allowedFields []string, defaultOrder string) SQLResult {
	var clauses []string
	var args []interface{}
	paramIndex := 1

	allowed := make(map[string]bool, len(allowedFields))
	for _, f := range allowedFields {
		allowed[f] = true
	}

	// ── exact match filters ──────────────────────────────────────────────
	for key, val := range fq.Filters {
		if !allowed[key] || val == nil {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("%s = $%d", key, paramIndex))
		args = append(args, val)
		paramIndex++
	}

	// ── search filters (LIKE / ILIKE) ────────────────────────────────────
	if fq.SearchFilters != nil {
		switch sf := fq.SearchFilters.(type) {
		case []SearchFilterItem:
			var searchClauses []string
			for _, item := range sf {
				if !allowed[item.Field] || item.Value == nil || item.Value == "" {
					continue
				}
				searchClauses = append(searchClauses, fmt.Sprintf("%s ILIKE $%d", item.Field, paramIndex))
				args = append(args, "%"+fmt.Sprintf("%v", item.Value)+"%")
				paramIndex++
			}
			if len(searchClauses) > 0 {
				clauses = append(clauses, "("+strings.Join(searchClauses, " OR ")+")")
			}
		case map[string]interface{}:
			var searchClauses []string
			for key, val := range sf {
				if !allowed[key] || val == nil || val == "" {
					continue
				}
				searchClauses = append(searchClauses, fmt.Sprintf("%s ILIKE $%d", key, paramIndex))
				args = append(args, "%"+fmt.Sprintf("%v", val)+"%")
				paramIndex++
			}
			if len(searchClauses) > 0 {
				clauses = append(clauses, "("+strings.Join(searchClauses, " OR ")+")")
			}
		}
	}

	// ── ranged filters (BETWEEN) ─────────────────────────────────────────
	for _, rf := range fq.RangedFilters {
		if !allowed[rf.Key] {
			continue
		}

		if rf.From != nil && rf.To != nil {
			// BETWEEN from AND to
			clauses = append(clauses, fmt.Sprintf("%s BETWEEN $%d AND $%d", rf.Key, paramIndex, paramIndex+1))
			args = append(args, rf.From, rf.To)
			paramIndex += 2
		} else if rf.From != nil {
			// >= from
			clauses = append(clauses, fmt.Sprintf("%s >= $%d", rf.Key, paramIndex))
			args = append(args, rf.From)
			paramIndex++
		} else if rf.To != nil {
			// <= to
			clauses = append(clauses, fmt.Sprintf("%s <= $%d", rf.Key, paramIndex))
			args = append(args, rf.To)
			paramIndex++
		}
	}

	// ── order ─────────────────────────────────────────────────────
	order := defaultOrder
	if fq.OrderKey != "" && allowed[fq.OrderKey] {
		rule := "DESC"
		if fq.OrderRule == "asc" {
			rule = "ASC"
		}
		order = fmt.Sprintf("%s %s", fq.OrderKey, rule)
	}

	// ── pagination ────────────────────────────────────────────────────
	limit := fq.Limit
	if limit <= 0 {
		limit = 10
	}
	limit = int(math.Min(math.Max(float64(limit), 1), 100))

	offset := (fq.Page - 1) * limit

	// ── build WHERE clause ────────────────────────────────────────────
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}

	return SQLResult{
		Where:  where,
		Args:   args,
		Order:  order,
		Limit:  limit,
		Offset: offset,
	}
}

// WrapPaginated wraps data with pagination metadata
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
