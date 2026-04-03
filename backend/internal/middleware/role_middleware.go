package middleware

import (
	"net/http"
	"trackly-backend/internal/service"
)

func RequireRole(role string, userService service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims := GetUserFromContext(r)
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userRole, err := userService.GetRole(r.Context(), claims.ID)
			if err != nil || userRole != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
