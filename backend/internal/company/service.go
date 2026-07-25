package company

import (
	"context"
	"trackly-backend/pkg/filter"
	"trackly-backend/pkg/httpx"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CompanyService struct {
	repo CompanyRepository
	log  *logrus.Logger
}

func NewCompanyService(repo CompanyRepository, log *logrus.Logger) *CompanyService {
	return &CompanyService{repo: repo, log: log}
}

func (s *CompanyService) GetAllCompanies(ctx context.Context, fq filter.FilteringQuery) (*filter.PaginatedResult[*CompanyResponse], *httpx.AppError) {
	data, total, err := s.repo.FindAll(ctx, fq)
	if err != nil {
		return nil, &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "Failed to retrieve companies",
		}
	}

	res := ToCompanyResponseList(data)
	result := filter.WrapPaginated(res, total, fq.Page, fq.Limit)

	return &result, nil
}

func (s *CompanyService) GetCompanyByID(ctx context.Context, id int) (*CompanyResponse, *httpx.AppError) {
	company, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, &httpx.AppError{
			Code:   httpx.ErrNotFound,
			Detail: "Company not found",
		}
	}

	return ToCompanyResponse(company), nil
}

func (s *CompanyService) CreateCompany(ctx context.Context, req *CreateCompanyRequest) (*CompanyResponse, *httpx.AppError) {
	company := &Company{
		Kode:              req.Kode,
		NamaPerusahaan:    req.NamaPerusahaan,
		TanggalPencatatan: ParseDatePointer(req.TanggalPencatatan),
		JumlahSaham:       req.JumlahSaham,
		PapanPencatatan:   req.PapanPencatatan,
	}

	created, err := s.repo.Create(ctx, company)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			return nil, &httpx.AppError{
				Code:   httpx.ErrConflict,
				Detail: "Company with the same kode already exists",
			}
		}
		return nil, &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "Failed to create company",
		}
	}

	return ToCompanyResponse(created), nil
}

func (s *CompanyService) UpdateCompany(ctx context.Context, id int, req *UpdateCompanyRequest) (*CompanyResponse, *httpx.AppError) {
	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			return nil, &httpx.AppError{
				Code:   httpx.ErrConflict,
				Detail: "Company with the same kode already exists",
			}
		}
		if err == gorm.ErrRecordNotFound {
			return nil, &httpx.AppError{
				Code:   httpx.ErrNotFound,
				Detail: "Company not found",
			}
		}
		if err == gorm.ErrInvalidData {
			return nil, &httpx.AppError{
				Code:   httpx.ErrValidation,
				Detail: "Invalid date format for TanggalPencatatan, expected YYYY-MM-DD",
			}
		}
		return nil, &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "Failed to update company",
		}
	}

	return ToCompanyResponse(updated), nil
}

func (s *CompanyService) DeleteCompany(ctx context.Context, id int) *httpx.AppError {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &httpx.AppError{
				Code:   httpx.ErrNotFound,
				Detail: "Company not found",
			}
		}
		return &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "Failed to delete company",
		}
	}

	return nil
}

func (s *CompanyService) Import(ctx context.Context, requests []CompanyImportRequest) *CompanyImportResponse {
	result := &CompanyImportResponse{
		TotalRows:     len(requests),
		SuccessCount:  0,
		FailureCount:  0,
		FailedRecords: []CompanyFailedRecord{},
	}

	var companies []Company

	for _, req := range requests {
		companies = append(companies, Company{
			Kode:              req.Kode,
			NamaPerusahaan:    req.NamaPerusahaan,
			TanggalPencatatan: ParseDatePointer(req.TanggalPencatatan),
			JumlahSaham:       req.JumlahSaham,
			PapanPencatatan:   req.PapanPencatatan,
		})
	}

	// Bulk create
	created, err := s.repo.BulkCreate(ctx, companies)
	if err != nil {
		if s.log != nil {
			s.log.Error(err)
		}
		result.FailureCount = len(requests)
		for i, req := range requests {
			result.FailedRecords = append(result.FailedRecords, CompanyFailedRecord{
				RowNumber: i + 2,
				Kode:      req.Kode,
				Reason:    err.Error(),
			})
		}
		return result
	}

	result.SuccessCount = len(created)
	return result
}
