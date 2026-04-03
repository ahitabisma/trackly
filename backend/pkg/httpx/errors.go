package httpx

import (
	"strings"
)

// ConstraintErrorMapping defines custom error messages for specific constraints
type ConstraintErrorMapping struct {
	ConstraintName  string
	FriendlyMessage string
}

// MapSQLError maps common SQL errors to user-friendly messages
// It supports generic patterns and can be extended with constraint-specific mappings
func MapSQLError(err error) string {
	return MapSQLErrorWithConstraints(err, nil)
}

// MapSQLErrorWithConstraints maps SQL errors with optional custom constraint mappings
// constraints: slice of ConstraintErrorMapping for application-specific constraints
func MapSQLErrorWithConstraints(err error, constraints []ConstraintErrorMapping) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Map duplicate key errors
	if strings.Contains(errMsg, "duplicate key value violates unique constraint") {
		// Check for custom constraint mappings first
		for _, c := range constraints {
			if strings.Contains(errMsg, c.ConstraintName) {
				return c.FriendlyMessage
			}
		}

		// Fallback to generic extraction from constraint name
		if idx := strings.Index(errMsg, "\""); idx != -1 {
			if endIdx := strings.Index(errMsg[idx+1:], "\""); endIdx != -1 {
				constraintName := errMsg[idx+1 : idx+1+endIdx]
				return "The value for " + constraintName + " already exists"
			}
		}
		return "This value already exists"
	}

	// Map foreign key errors
	if strings.Contains(errMsg, "violates foreign key constraint") {
		// Check for custom constraint mappings
		for _, c := range constraints {
			if strings.Contains(errMsg, c.ConstraintName) {
				return c.FriendlyMessage
			}
		}
		return "Referenced data not found"
	}

	// Map not null errors
	if strings.Contains(errMsg, "violates not-null constraint") {
		// Check for custom constraint mappings
		for _, c := range constraints {
			if strings.Contains(errMsg, c.ConstraintName) {
				return c.FriendlyMessage
			}
		}

		// Extract column name from error message
		if strings.Contains(errMsg, "\"") {
			if idx := strings.LastIndex(errMsg, "\""); idx != -1 {
				if startIdx := strings.LastIndex(errMsg[:idx], "\""); startIdx != -1 {
					columnName := errMsg[startIdx+1 : idx]
					return "The field " + columnName + " is required"
				}
			}
		}
		return "Required field is missing"
	}

	// Map check constraint errors
	if strings.Contains(errMsg, "violates check constraint") {
		// Check for custom constraint mappings
		for _, c := range constraints {
			if strings.Contains(errMsg, c.ConstraintName) {
				return c.FriendlyMessage
			}
		}
		return "Value does not meet required criteria"
	}

	// Map unique constraint errors (similar to duplicate key)
	if strings.Contains(errMsg, "violates unique constraint") {
		// Check for custom constraint mappings
		for _, c := range constraints {
			if strings.Contains(errMsg, c.ConstraintName) {
				return c.FriendlyMessage
			}
		}
		return "This value must be unique"
	}

	// Default: return generic message
	return "An error occurred while processing your request"
}
