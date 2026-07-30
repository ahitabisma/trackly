package screening

import (
	"net/http"

	"trackly-backend/pkg/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetLatest(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.repo.GetLatest(r.Context())
	if err != nil {
		httpx.WriteJSON(w, r, http.StatusInternalServerError, httpx.Error(httpx.ErrInternal, "fetch latest screening", err.Error()))
		return
	}
	if results == nil {
		results = []DailyScreeningResult{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, httpx.Success(results, "ok"))
}

func (h *Handler) GetByDate(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	results, err := h.svc.repo.GetByDate(r.Context(), date)
	if err != nil {
		httpx.WriteJSON(w, r, http.StatusInternalServerError, httpx.Error(httpx.ErrInternal, "fetch screening", err.Error()))
		return
	}
	if results == nil {
		results = []DailyScreeningResult{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, httpx.Success(results, "ok"))
}

func (h *Handler) TriggerScreening(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := h.svc.RunNightlyScreening(r.Context()); err != nil {
			h.svc.log.WithError(err).Error("manual screening trigger failed")
		}
	}()
	httpx.WriteJSON(w, r, http.StatusAccepted, httpx.Success(map[string]string{"status": "screening started"}, "ok"))
}
