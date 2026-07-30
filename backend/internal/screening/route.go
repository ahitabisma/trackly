package screening

import "net/http"

func SetupScreeningRoutes(mux *http.ServeMux, handler *Handler, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /api/screening/latest", authMiddleware(http.HandlerFunc(handler.GetLatest)))
	mux.Handle("GET /api/screening/{date}", authMiddleware(http.HandlerFunc(handler.GetByDate)))
	mux.Handle("POST /api/screening/trigger", authMiddleware(http.HandlerFunc(handler.TriggerScreening)))
}
