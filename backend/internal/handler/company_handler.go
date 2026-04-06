package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"trackly-backend/internal/dto"
	"trackly-backend/internal/service"
	"trackly-backend/pkg/filter"
	"trackly-backend/pkg/httpx"
	"trackly-backend/pkg/parser"
)

type CompanyHandler struct {
	service *service.CompanyService
}

func NewCompanyHandler(s *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{s}
}

// parseCompanyCSV is a helper function that wraps the CSV parser
func parseCompanyCSV(file interface{}) ([]dto.CompanyImportRequest, error) {
	// Convert file to io.Reader
	readCloser, ok := file.(interface {
		Read([]byte) (int, error)
	})
	if !ok {
		return nil, fmt.Errorf("invalid file type")
	}

	return parser.ParseCompanyCSV(readCloser)
}

func (h *CompanyHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	fq := filter.ParseFromRequest(r)

	data, err := h.service.GetAll(r.Context(), fq)
	if err != nil {
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	respMap := data.(map[string]interface{})
	companies := respMap["data"].([]dto.CompanyResponse)
	total := respMap["total"].(int64)

	// Wrap paginated result
	paginatedResult := filter.WrapPaginated(companies, total, fq.Page, fq.Limit)

	httpx.SuccessWithPagination(w, paginatedResult.Data, httpx.Pagination(paginatedResult.Pagination), "Success get companies")
	return nil
}

// GetByID retrieves a single company by ID
func (h *CompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return httpx.NewBadRequestError("Invalid company ID")
	}

	company, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "company not found" {
			return httpx.NewNotFoundError("Company not found")
		}
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	httpx.Success(w, company, "Success get company detail")
	return nil
}

// Create creates a new company
func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var req dto.CreateCompanyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.NewBadRequestError("Invalid request body")
	}

	// Validate required fields
	fieldErrors := []httpx.FieldError{}
	if req.Kode == "" {
		fieldErrors = append(fieldErrors, httpx.NewFieldError("kode", "Kode is required"))
	}
	if req.NamaPerusahaan == "" {
		fieldErrors = append(fieldErrors, httpx.NewFieldError("nama_perusahaan", "Nama perusahaan is required"))
	}

	if len(fieldErrors) > 0 {
		return httpx.NewValidationError("Validation failed", fieldErrors)
	}

	company, err := h.service.Create(r.Context(), &req)
	if err != nil {
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	httpx.Created(w, company, "Company created successfully")
	return nil
}

// Update updates an existing company
func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return httpx.NewBadRequestError("Invalid company ID")
	}

	var req dto.UpdateCompanyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.NewBadRequestError("Invalid request body")
	}

	// Check if company exists
	_, err = h.service.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "company not found" {
			return httpx.NewNotFoundError("Company not found")
		}
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	company, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	httpx.Success(w, company, "Company updated successfully")
	return nil
}

// Delete deletes a company
func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return httpx.NewBadRequestError("Invalid company ID")
	}

	// Check if company exists
	_, err = h.service.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "company not found" {
			return httpx.NewNotFoundError("Company not found")
		}
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		friendlyMsg := httpx.MapSQLError(err)
		return httpx.NewCustomError(http.StatusInternalServerError, "INTERNAL_ERROR", friendlyMsg)
	}

	httpx.Success(w, nil, "Company deleted successfully")
	return nil
}

// Import imports companies from CSV file
func (h *CompanyHandler) Import(w http.ResponseWriter, r *http.Request) error {
	// Parse multipart form with max 10MB file size
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return httpx.NewBadRequestError("Failed to parse form data")
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return httpx.NewBadRequestError("File is required")
	}
	defer file.Close()

	// Parse CSV
	importRequests, err := parseCompanyCSV(file)
	if err != nil {
		return httpx.NewBadRequestError(fmt.Sprintf("Failed to parse CSV: %v", err))
	}

	if len(importRequests) == 0 {
		return httpx.NewBadRequestError("CSV file is empty")
	}

	// Import companies
	result := h.service.Import(r.Context(), importRequests)

	httpx.Success(w, result, "Import completed")
	return nil
}
