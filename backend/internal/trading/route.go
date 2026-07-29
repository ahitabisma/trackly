package trading

import "net/http"

func SetupTradingRoutes(mux *http.ServeMux, handler *TradingHandler, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("POST /api/transactions", authMiddleware(http.HandlerFunc(handler.PostTransaction)))
	mux.Handle("GET /api/transactions", authMiddleware(http.HandlerFunc(handler.GetTransactions)))
	mux.Handle("GET /api/positions", authMiddleware(http.HandlerFunc(handler.GetPositions)))
	mux.Handle("GET /api/positions/{ticker}/analysis", authMiddleware(http.HandlerFunc(handler.GetPositionAnalysis)))
	mux.Handle("PATCH /api/transactions/{id}", authMiddleware(http.HandlerFunc(handler.UpdateTransaction)))
	mux.Handle("DELETE /api/transactions/{id}", authMiddleware(http.HandlerFunc(handler.DeleteTransaction)))
}
