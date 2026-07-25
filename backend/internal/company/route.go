package company

import "net/http"

func SetupCompanyRoutes(
	mux *http.ServeMux,
	handler *CompanyHandler,
	authMiddleware func(http.Handler) http.Handler,
	adminMiddleware func(http.Handler) http.Handler,
) {
	mux.Handle("GET /companies", http.HandlerFunc(handler.ListCompanies))
	mux.Handle("GET /companies/", http.HandlerFunc(handler.GetCompany))

	mux.Handle("POST /companies", adminMiddleware(http.HandlerFunc(handler.CreateCompany)))
	mux.Handle("PUT /companies/", adminMiddleware(http.HandlerFunc(handler.UpdateCompany)))
	mux.Handle("DELETE /companies/", adminMiddleware(http.HandlerFunc(handler.DeleteCompany)))
	mux.Handle("POST /companies/import", adminMiddleware(http.HandlerFunc(handler.ImportCompanies)))
}
