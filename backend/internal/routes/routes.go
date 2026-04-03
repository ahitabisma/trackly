package routes

import (
	"net/http"
	"trackly-backend/internal/handler"
	"trackly-backend/pkg/httpx"
)

func Setup(mux *http.ServeMux, companyHandler *handler.CompanyHandler) {
	setupCompanyRoutes(mux, companyHandler)
}

func setupCompanyRoutes(mux *http.ServeMux, h *handler.CompanyHandler) {
	mux.HandleFunc("GET /companies", httpx.Wrap(h.GetAll))
	mux.HandleFunc("GET /companies/{id}", httpx.Wrap(h.GetByID))
	mux.HandleFunc("POST /companies", httpx.Wrap(h.Create))
	mux.HandleFunc("PUT /companies/{id}", httpx.Wrap(h.Update))
	mux.HandleFunc("DELETE /companies/{id}", httpx.Wrap(h.Delete))
	mux.HandleFunc("POST /companies/import", httpx.Wrap(h.Import))
}
