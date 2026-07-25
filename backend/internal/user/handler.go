package user

import (
	"net/http"
	"strconv"

	"trackly-backend/pkg/filter"
	"trackly-backend/pkg/httpx"
	customLogger "trackly-backend/pkg/logger"

	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	service *UserService
	log     *logrus.Logger
}

func NewUserHandler(service *UserService, log *logrus.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		log:     log,
	}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		resp := httpx.Error(httpx.ErrValidation, "User id not found in context", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusUnauthorized, resp)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid user id format", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"parse_error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusUnauthorized, resp)
		return
	}

	res, appErr := h.service.GetUserProfile(r.Context(), uint(userID))
	if appErr != nil {
		if appErr.Code == httpx.ErrInvalidCredentials {
			resp := httpx.Error(appErr.Code, "User not found", appErr.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"user_id": userID,
			})
			httpx.WriteJSON(w, r, http.StatusNotFound, resp)
			return
		}

		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"user_id": userID,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(res, "User profile retrieved successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"user_id": userID,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	fq := filter.ParseFromRequest(r)

	res, appErr := h.service.GetAllUsers(r.Context(), fq)
	if appErr != nil {
		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.SuccessWithPagination(res.Data, "Users retrieved successfully", int(res.Pagination.Total), res.Pagination.Page, res.Pagination.Limit)
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
