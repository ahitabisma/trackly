package auth

import "net/http"

func SetupAuthRoutes(mux *http.ServeMux, handler *AuthHandler) {
	mux.HandleFunc("POST /auth/register", handler.Register)
	mux.HandleFunc("POST /auth/login", handler.Login)
}
