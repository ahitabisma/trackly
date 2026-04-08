package user

import "net/http"

func SetupUserRoutes(
	mux *http.ServeMux,
	handler *UserHandler,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) {
	// Public routes (protected by auth)
	mux.Handle("GET /me", authMiddleware(http.HandlerFunc(handler.GetProfile)))

	// Admin-only routes (protected by auth + role middleware)
	mux.Handle("GET /users", adminMiddleware(http.HandlerFunc(handler.ListUsers)))
}
