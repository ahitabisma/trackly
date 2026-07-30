package trading

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"trackly-backend/pkg/httpx"
	customLogger "trackly-backend/pkg/logger"
	"trackly-backend/pkg/validatorx"

	"github.com/sirupsen/logrus"
)

type TradingHandler struct {
	svc *TradingService
	log *logrus.Logger
}

func NewTradingHandler(svc *TradingService, log *logrus.Logger) *TradingHandler {
	return &TradingHandler{svc: svc, log: log}
}

func userIDFromContext(r *http.Request) uint {
	id, _ := strconv.ParseUint(r.Header.Get("X-User-ID"), 10, 64)
	return uint(id)
}

func (h *TradingHandler) PostTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid request body", err.Error())
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}
	if fields := validatorx.ValidateStruct(req); fields != nil {
		httpx.WriteJSON(w, r, http.StatusBadRequest, httpx.ValidationError(fields))
		return
	}

	userID := userIDFromContext(r)
	txn, err := h.svc.CreateTransaction(r.Context(), &req, userID)
	if err != nil {
		resp := httpx.Error(httpx.ErrInternal, "Failed to create transaction", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, nil)
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(txn, "Transaction created")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"ticker": req.Ticker})
	httpx.WriteJSON(w, r, http.StatusCreated, resp)
}

func (h *TradingHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
	userID := userIDFromContext(r)

	txns, err := h.svc.GetTransactions(r.Context(), userID, ticker)
	if err != nil {
		resp := httpx.Error(httpx.ErrInternal, "Failed to get transactions", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, nil)
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(txns, "Transactions retrieved")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"count": len(txns)})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *TradingHandler) GetPositions(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	positions, err := h.svc.GetAllOpenPositions(r.Context(), userID)
	if err != nil {
		resp := httpx.Error(httpx.ErrInternal, "Failed to get positions", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, nil)
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(positions, "Positions retrieved")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"count": len(positions)})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *TradingHandler) GetPositionAnalysis(w http.ResponseWriter, r *http.Request) {
	ticker := r.PathValue("ticker")
	if ticker == "" {
		resp := httpx.Error(httpx.ErrValidation, "Ticker is required", "")
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	userID := userIDFromContext(r)
	result, err := h.svc.GetPositionAnalysis(r.Context(), userID, ticker)
	if err != nil {
		if errors.Is(err, ErrNoTransactions) || errors.Is(err, ErrPositionClosed) {
			resp := httpx.Error(httpx.ErrNotFound, err.Error(), "")
			httpx.WriteJSON(w, r, http.StatusNotFound, resp)
			return
		}
		resp := httpx.Error(httpx.ErrInternal, "Position analysis failed", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{"ticker": ticker})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(result, "Position analysis completed")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{"ticker": ticker})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *TradingHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteJSON(w, r, http.StatusBadRequest, httpx.Error(httpx.ErrValidation, "ID is required", ""))
		return
	}

	defer r.Body.Close()
	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, r, http.StatusBadRequest, httpx.Error(httpx.ErrValidation, "Invalid request body", err.Error()))
		return
	}

	userID := userIDFromContext(r)
	txn, err := h.svc.UpdateTransaction(r.Context(), id, userID, &req)
	if err != nil {
		httpx.WriteJSON(w, r, http.StatusInternalServerError, httpx.Error(httpx.ErrInternal, "Failed to update transaction", err.Error()))
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, httpx.Success(txn, "Transaction updated"))
}

func (h *TradingHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		httpx.WriteJSON(w, r, http.StatusBadRequest, httpx.Error(httpx.ErrValidation, "ID is required", ""))
		return
	}

	userID := userIDFromContext(r)
	if err := h.svc.DeleteTransaction(r.Context(), id, userID); err != nil {
		httpx.WriteJSON(w, r, http.StatusInternalServerError, httpx.Error(httpx.ErrInternal, "Failed to delete transaction", err.Error()))
		return
	}

	httpx.WriteJSON(w, r, http.StatusOK, httpx.Success(nil, "Transaction deleted"))
}


