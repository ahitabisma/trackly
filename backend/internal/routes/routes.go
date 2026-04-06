package routes

import (
	"net/http"

	"trackly-backend/internal/handler"
	"trackly-backend/internal/middleware"
	"trackly-backend/internal/service"
	"trackly-backend/pkg/httpx"
)

func Setup(
	mux *http.ServeMux,
	companyHandler *handler.CompanyHandler,
	authHandler *handler.AuthHandler,
	jwtSecret string,
	userService service.UserService,
) {

	// middleware
	auth := middleware.AuthMiddleware(jwtSecret)
	admin := middleware.RequireRole("admin", userService)

	// ── PUBLIC ─────────────────────────────
	mux.HandleFunc("POST /login", httpx.Wrap(authHandler.Login))
	mux.HandleFunc("GET /companies", httpx.Wrap(companyHandler.GetAll))
	mux.HandleFunc("GET /companies/{id}", httpx.Wrap(companyHandler.GetByID))

	// ── PROTECTED (ADMIN) ──────────────────
	mux.Handle(
		"POST /companies",
		auth(admin(http.HandlerFunc(httpx.Wrap(companyHandler.Create)))),
	)

	mux.Handle(
		"PUT /companies/{id}",
		auth(admin(http.HandlerFunc(httpx.Wrap(companyHandler.Update)))),
	)

	mux.Handle(
		"DELETE /companies/{id}",
		auth(admin(http.HandlerFunc(httpx.Wrap(companyHandler.Delete)))),
	)

	mux.Handle(
		"POST /companies/import",
		auth(admin(http.HandlerFunc(httpx.Wrap(companyHandler.Import)))),
	)
}
