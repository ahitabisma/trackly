package company

import (
	"encoding/json"
	"net/http"
	"trackly-backend/pkg/filter"
	"trackly-backend/pkg/httpx"
	customLogger "trackly-backend/pkg/logger"
	"trackly-backend/pkg/validatorx"

	"github.com/sirupsen/logrus"
)

type CompanyHandler struct {
	svc *CompanyService
	log *logrus.Logger
}

func NewCompanyHandler(svc *CompanyService, log *logrus.Logger) *CompanyHandler {
	return &CompanyHandler{
		svc: svc,
		log: log,
	}
}

func (h *CompanyHandler) ListCompanies(w http.ResponseWriter, r *http.Request) {
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

	res, appErr := h.svc.GetAllCompanies(r.Context(), fq)
	if appErr != nil {
		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.SuccessWithPagination(res.Data, "Companies retrieved successfully", int(res.Pagination.Total), res.Pagination.Page, res.Pagination.Limit)
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *CompanyHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	id, err := ParseIDFromURL(r.URL.Path)
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid company ID", "ID must be a valid integer")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"parse_error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	res, appErr := h.svc.GetCompanyByID(r.Context(), id)
	if appErr != nil {
		if appErr.Code == httpx.ErrNotFound {
			resp := httpx.Error(appErr.Code, "Company not found", appErr.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"company_id": id,
			})
			httpx.WriteJSON(w, r, http.StatusNotFound, resp)
			return
		}

		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"company_id": id,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(res, "Company retrieved successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"company_id": id,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	var req CreateCompanyRequest
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
			"request": req,
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	res, appErr := h.svc.CreateCompany(r.Context(), &req)
	if appErr != nil {
		if appErr.Code == httpx.ErrConflict {
			resp := httpx.Error(appErr.Code, "Company already exists", appErr.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"request": req,
			})
			httpx.WriteJSON(w, r, http.StatusConflict, resp)
			return
		}

		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"request": req,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(res, "Company created successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"company_id": res.ID,
	})
	httpx.WriteJSON(w, r, http.StatusCreated, resp)
}

func (h *CompanyHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	id, err := ParseIDFromURL(r.URL.Path)
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid company ID", "ID must be a valid integer")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"parse_error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	var req UpdateCompanyRequest
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
			"request": req,
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	res, appErr := h.svc.UpdateCompany(r.Context(), id, &req)
	if appErr != nil {
		if appErr.Code == httpx.ErrNotFound {
			resp := httpx.Error(appErr.Code, "Company not found", appErr.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"company_id": id,
			})
			httpx.WriteJSON(w, r, http.StatusNotFound, resp)
			return
		}
		if appErr.Code == httpx.ErrValidation {
			resp := httpx.Error(appErr.Code, "Validation error", appErr.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"company_id": id,
				"request":    req,
			})
			httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
			return
		}

		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"company_id": id,
			"request":    req,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(res, "Company updated successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"company_id": id,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *CompanyHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	id, err := ParseIDFromURL(r.URL.Path)
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid company ID", "ID must be a valid integer")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"parse_error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	appErr := h.svc.DeleteCompany(r.Context(), id)
	if appErr != nil {
		if appErr.Code == httpx.ErrNotFound {
			resp := httpx.Error(appErr.Code, "Company not found", appErr.Detail)
			customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
				"company_id": id,
			})
			httpx.WriteJSON(w, r, http.StatusNotFound, resp)
			return
		}

		resp := httpx.Error(appErr.Code, "Something went wrong", appErr.Detail)
		customLogger.LogHTTPInternalError(h.log, resp, map[string]interface{}{
			"company_id": id,
		})
		httpx.WriteJSON(w, r, http.StatusInternalServerError, resp)
		return
	}

	resp := httpx.Success(nil, "Company deleted successfully")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"company_id": id,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (h *CompanyHandler) ImportCompanies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := httpx.Error(httpx.ErrValidation, "Method not allowed", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"method": r.Method,
		})
		httpx.WriteJSON(w, r, http.StatusMethodNotAllowed, resp)
		return
	}

	defer r.Body.Close()

	// Parse multipart form with max file size 10MB
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid request", err.Error())
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
	if header.Header.Get("Content-Type") != "text/csv" && header.Filename[len(header.Filename)-4:] != ".csv" {
		resp := httpx.Error(httpx.ErrValidation, "File must be CSV", "")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"content_type": header.Header.Get("Content-Type"),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	// Parse CSV using parser
	requests, err := ParseCompanyCSV(file)
	if err != nil {
		resp := httpx.Error(httpx.ErrValidation, "Invalid CSV format", err.Error())
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{
			"error": err.Error(),
		})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	if len(requests) == 0 {
		resp := httpx.Error(httpx.ErrValidation, "CSV is empty", "At least one data row required")
		customLogger.LogHTTPError(h.log, resp, map[string]interface{}{})
		httpx.WriteJSON(w, r, http.StatusBadRequest, resp)
		return
	}

	// Call service
	importRes := h.svc.Import(r.Context(), requests)

	resp := httpx.Success(importRes, "Import completed")
	customLogger.LogHTTPSuccess(h.log, resp, map[string]interface{}{
		"total_rows":    importRes.TotalRows,
		"success_count": importRes.SuccessCount,
		"failure_count": importRes.FailureCount,
	})
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
