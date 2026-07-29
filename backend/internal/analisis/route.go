package analisis

import "net/http"

func SetupAnalisisRoutes(mux *http.ServeMux, handler *AnalisisHandler) {
	mux.HandleFunc("GET /api/tickers", handler.SearchTickers)
	mux.HandleFunc("GET /api/ticker/{kode}", handler.GetTicker)
	mux.HandleFunc("POST /api/analisis", handler.PostAnalisis)
	mux.HandleFunc("POST /api/analisis/ai-insight", handler.PostAiInsight)
}
