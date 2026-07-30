package analisis

import "net/http"

func SetupAnalisisRoutes(mux *http.ServeMux, handler *AnalisisHandler, authMiddleware func(http.Handler) http.Handler) {
	mux.HandleFunc("GET /api/tickers", handler.SearchTickers)
	mux.HandleFunc("GET /api/ticker/{kode}", handler.GetTicker)
	mux.Handle("POST /api/analisis", authMiddleware(http.HandlerFunc(handler.PostAnalisis)))
	mux.Handle("POST /api/analisis/ai-insight", authMiddleware(http.HandlerFunc(handler.PostAiInsight)))
}
