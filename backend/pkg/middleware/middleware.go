package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"trackly-backend/internal/auth"
	"trackly-backend/pkg/httpx"
	customLogger "trackly-backend/pkg/logger"

	"github.com/sirupsen/logrus"
)

func AuthMiddleware(jwtSvc *auth.JwtService, log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				resp := httpx.Error(httpx.ErrValidation, "missing authorization header", "")
				customLogger.LogHTTPError(log, resp, map[string]interface{}{
					"path": r.RequestURI,
				})
				httpx.WriteJSON(w, r, http.StatusUnauthorized, resp)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				resp := httpx.Error(httpx.ErrValidation, "invalid authorization header format", "")
				customLogger.LogHTTPError(log, resp, map[string]interface{}{
					"path": r.RequestURI,
				})
				httpx.WriteJSON(w, r, http.StatusUnauthorized, resp)
				return
			}

			claims, err := jwtSvc.ValidateToken(parts[1])
			if err != nil {
				resp := httpx.Error(httpx.ErrValidation, "invalid or expired token", err.Error())
				customLogger.LogHTTPError(log, resp, map[string]interface{}{
					"path": r.RequestURI,
				})
				httpx.WriteJSON(w, r, http.StatusUnauthorized, resp)
				return
			}

			r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.Sub))
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)

			log.WithFields(logrus.Fields{
				"user_id":    claims.Sub,
				"user_email": claims.Email,
				"path":       r.RequestURI,
			}).Debug("user authenticated successfully")

			next.ServeHTTP(w, r)
		})
	}
}

func RoleMiddleware(requiredRoles ...string) func(*logrus.Logger) func(http.Handler) http.Handler {
	return func(log *logrus.Logger) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userRole := r.Header.Get("X-User-Role")
				userID := r.Header.Get("X-User-ID")

				isAuthorized := false
				for _, role := range requiredRoles {
					if userRole == role {
						isAuthorized = true
						break
					}
				}

				if !isAuthorized {
					resp := httpx.Error(httpx.ErrForbidden, "Insufficient permissions", "User is not authorized")
					customLogger.LogHTTPError(log, resp, map[string]interface{}{
						"user_id":        userID,
						"user_role":      userRole,
						"required_roles": requiredRoles,
						"path":           r.RequestURI,
					})
					httpx.WriteJSON(w, r, http.StatusForbidden, resp)
					return
				}

				log.WithFields(logrus.Fields{
					"user_id":   userID,
					"user_role": userRole,
					"path":      r.RequestURI,
				}).Debug("user authorized by role")

				next.ServeHTTP(w, r)
			})
		}
	}
}

func ChainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
