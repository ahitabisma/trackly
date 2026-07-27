package analisis

import (
	"encoding/json"
	"errors"
	"net/http"
	"trackly-backend/pkg/httpx"
	customLogger "trackly-backend/pkg/logger"
	"trackly-backend/pkg/validatorx"

	"github.com/sirupsen/logrus"
)

type AnalisisHandler struct {
	svc *AnalisisService
	log *logrus.Logger
}

func NewAnalisisHandler(svc *AnalisisService, log *logrus.Logger) *AnalisisHandler {
	return &AnalisisHandler{
		svc: svc,
		log: log,
	}
}

func (h *AnalisisHandler) SearchTickers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	results, err := h.svc.SearchTickers(r.Context())
	if err != nil {
		resp := httpx.Error(httpx.ErrInternal, "Failed to search tickers", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, nil)
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	if results == nil {
		results = []TickerSearchResult{}
	}

	resp := httpx.Success(results, "Tickers retrieved successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"count": len(results)})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *AnalisisHandler) GetTicker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	kode := r.PathValue("kode")
	if kode == "" {
		resp := httpx.Error(httpx.ErrValidation, "Kode is required", "")
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	snap, err := h.svc.GetTicker(r.Context(), kode)
	if err != nil {
		resp := httpx.Error(httpx.ErrNotFound, "Ticker not found", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{"kode": kode})
		httpx.WriteJSON(w, r, http.StatusNotFound, resp)
		return
	}

	resp := httpx.Success(snap, "Ticker retrieved successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"kode": kode})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *AnalisisHandler) PostAnalisis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	var req AnalisisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid request body", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	if fields := validatorx.ValidateStruct(req); fields != nil {
		resp := httpx.ValidationError(fields)
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	result, err := h.svc.RunAnalisis(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrWorkerPoolBusy) {
			resp := httpx.Error(httpx.ErrTooManyRequests, "Server sedang sibuk, coba lagi sebentar lagi", "")
			httpx.WriteJSON(w, r, http.StatusServiceUnavailable, resp)
			return
		}
		resp := httpx.Error(httpx.ErrInternal, "Analisis failed", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{"ticker": req.Ticker})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(result, "Analisis completed")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"ticker": req.Ticker})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
