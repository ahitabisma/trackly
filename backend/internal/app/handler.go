package app

import (
	"net/http"
	"time"
	"trackly-backend/pkg/config"
	"trackly-backend/pkg/httpx"

	customLogger "trackly-backend/pkg/logger"

	"github.com/sirupsen/logrus"
)

type AppHandler struct {
	log *logrus.Logger
	cfg *config.Config
}

func NewAppHandler(log *logrus.Logger, cfg *config.Config) *AppHandler {
	return &AppHandler{
		log: log,
		cfg: cfg,
	}
}

func (h *AppHandler) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	data := map[string]interface{}{
		"app_name":    h.cfg.App.Name,
		"environment": h.cfg.App.Env,
		"ip_address":  ip,
		"user_agent":  r.UserAgent(),
		"server_time": time.Now().Format(time.RFC3339),
		"status":      "running",
	}

	resp := httpx.Success(data, "Welcome to Trackly API!")
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
