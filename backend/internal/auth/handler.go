package auth

import (
	"encoding/json"
	"net/http"

	"trackly-backend/pkg/httpx"
	customLogger "trackly-backend/pkg/logger"
	"trackly-backend/pkg/validatorx"

	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	service *AuthService
	log     *logrus.Logger
}

func NewAuthHandler(service *AuthService, log *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		log:     log,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid request body", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"decode_error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	if fields := validatorx.ValidateStruct(req); fields != nil {
		resp := httpx.ValidationError(fields)
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"email": req.Email,
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	if err := h.service.Register(&req); err != nil {
		if err.Code == httpx.ErrValidation {
			resp := httpx.ValidationError(err.Fields)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"email": req.Email,
			})
			httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
			return
		}

		resp := httpx.Error(err.Code, "Something went wrong", err.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"email": req.Email,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(nil, "User created successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"email": req.Email,
	})
	httpx.WriteJSON(w, r, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid request body", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"decode_error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	if fields := validatorx.ValidateStruct(req); fields != nil {
		resp := httpx.ValidationError(fields)
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"email": req.Email,
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	res, err := h.service.Login(&req)
	if err != nil {
		if err.Code == httpx.ErrInvalidCredentials {
			resp := httpx.Error(err.Code, "Invalid email or password", err.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"email": req.Email,
			})
			httpx.WriteJSON(w, r, http.StatusUnauthorized, resp)
			return
		}

		resp := httpx.Error(err.Code, "Something went wrong", err.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"email": req.Email,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(res, "Login successful")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"email": req.Email,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
