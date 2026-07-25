package shareholding

import "net/http"

func SetupShareholdingRoutes(
	mux *http.ServeMux,
	handler *ShareholdingHandler,
	adminMiddleware func(http.Handler) http.Handler,
) {
	mux.Handle("GET /shareholdings", http.HandlerFunc(handler.ListShareholdings))
	mux.Handle("POST /shareholdings/import", adminMiddleware(http.HandlerFunc(handler.ImportShareholdings)))
	mux.Handle("GET /shareholdings/import/jobs/{jobID}", http.HandlerFunc(handler.GetImportJobStatus))
}
