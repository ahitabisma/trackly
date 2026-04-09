package shareholding

import (
	"net/http"
	"trackly-backend/pkg/httpx"
	"trackly-backend/pkg/jobs"
	customLogger "trackly-backend/pkg/logger"

	"github.com/sirupsen/logrus"
)

type ShareholdingHandler struct {
	svc   *ShareHoldingService
	queue jobs.Queue
	log   *logrus.Logger
}

func NewShareholdingHandler(svc *ShareHoldingService, log *logrus.Logger) *ShareholdingHandler {
	return &ShareholdingHandler{
		svc: svc,
		log: log,
	}
}

func NewShareholdingHandlerWithQueue(svc *ShareHoldingService, queue jobs.Queue, log *logrus.Logger) *ShareholdingHandler {
	return &ShareholdingHandler{
		svc:   svc,
		queue: queue,
		log:   log,
	}
}

func (h *ShareholdingHandler) ImportShareholdings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	// Parse multipart form with max file size 50MB
	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Failed to parse form", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "File is required", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}
	defer file.Close()

	// Verify file is CSV
	if header.Header.Get("Content-Type") != "text/csv" && len(header.Filename) > 4 && header.Filename[len(header.Filename)-4:] != ".csv" {
		resp := httpx.Error(httpx.ErrValidation, "File must be CSV", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"content_type": header.Header.Get("Content-Type"),
			"filename":     header.Filename,
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	// Parse CSV
	rows, err := ParseShareholdingCSV(file)
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid CSV format", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	if len(rows) == 0 {
		resp := httpx.Error(httpx.ErrValidation, "CSV is empty", "At least one data row required")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	// If queue is available, use job processing; otherwise, process synchronously
	if h.queue != nil {
		h.importWithQueue(w, r, rows)
		return
	}

	h.importSynchronous(w, r, rows)
}

// importWithQueue queues the import job and returns job ID
func (h *ShareholdingHandler) importWithQueue(w http.ResponseWriter, r *http.Request, rows []ShareHoldingImportRow) {
	payload := ShareholdingImportPayload{
		Rows: rows,
	}

	jobID, err := h.queue.Enqueue(r.Context(), jobs.JobTypeShareholdingImport, payload)
	if err != nil {
		resp := httpx.Error(httpx.ErrInternal, "Failed to queue job", err.Error())
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	response := map[string]interface{}{
		"job_id":     jobID,
		"status":     "pending",
		"total_rows": len(rows),
		"message":    "Import job queued successfully",
	}

	resp := httpx.Success(response, "Import job queued")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"job_id":     jobID,
		"total_rows": len(rows),
	})
	httpx.WriteJSON(w, r, http.StatusAccepted, resp)
}

// importSynchronous processes import immediately (fallback)
func (h *ShareholdingHandler) importSynchronous(w http.ResponseWriter, r *http.Request, rows []ShareHoldingImportRow) {
	importRes, appErr := h.svc.Import(r.Context(), rows, 500)
	if appErr != nil {
		resp := httpx.Error(appErr.Code, "Import failed", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"total_rows": importRes.TotalRows,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(importRes, "Import completed")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"total_rows":    importRes.TotalRows,
		"inserted":      importRes.Inserted,
		"updated":       importRes.Updated,
		"skipped":       importRes.SkippedInvalid,
		"new_investors": importRes.NewInvestors,
		"new_companies": importRes.NewCompanies,
		"error_count":   len(importRes.Errors),
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

// GetImportJobStatus retrieves the status of an import job
func (h *ShareholdingHandler) GetImportJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	if h.queue == nil {
		resp := httpx.Error(httpx.ErrValidation, "Job queue not available", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	// Extract job ID from URL path
	jobID := r.PathValue("jobID")
	if jobID == "" {
		resp := httpx.Error(httpx.ErrValidation, "Job ID is required", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	job, err := h.queue.GetJob(r.Context(), jobID)
	if err != nil {
		resp := httpx.Error(httpx.ErrInternal, "Failed to get job status", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"job_id": jobID,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	if job == nil {
		resp := httpx.Error(httpx.ErrNotFound, "Job not found", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"job_id": jobID,
		})
		httpx.WriteJSON(w, r, http.StatusNotFound, resp)
		return
	}

	resp := httpx.Success(job, "Job status retrieved")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"job_id": jobID,
		"status": job.Status,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
