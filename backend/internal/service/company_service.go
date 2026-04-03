package service

import (
	"context"
	"fmt"
	"time"

	"trackly-backend/internal/dto"
	"trackly-backend/internal/model"
	"trackly-backend/internal/repository"
	"trackly-backend/pkg/filter"

	"github.com/sirupsen/logrus"
)

type CompanyService struct {
	repo repository.CompanyRepository
	log  *logrus.Logger
}

func NewCompanyService(r repository.CompanyRepository, log *logrus.Logger) *CompanyService {
	return &CompanyService{repo: r, log: log}
}

func (s *CompanyService) GetAll(ctx context.Context, fq filter.FilteringQuery) (interface{}, error) {

	data, total, err := s.repo.FindAll(ctx, fq)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	res := dto.ToCompanyResponseList(data)

	return map[string]interface{}{
		"data":  res,
		"total": total,
	}, nil
}

func (s *CompanyService) GetByID(ctx context.Context, id int) (*dto.CompanyResponse, error) {
	company, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	return dto.ToCompanyResponse(company), nil
}

func (s *CompanyService) Create(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
	var tanggalPencatatan *time.Time
	if req.TanggalPencatatan != nil {
		t, err := time.Parse("2006-01-02", *req.TanggalPencatatan)
		if err != nil {
			return nil, fmt.Errorf("invalid tanggal_pencatatan format, use YYYY-MM-DD")
		}
		tanggalPencatatan = &t
	}

	company := &model.Company{
		Kode:              req.Kode,
		NamaPerusahaan:    req.NamaPerusahaan,
		TanggalPencatatan: tanggalPencatatan,
		JumlahSaham:       req.JumlahSaham,
		PapanPencatatan:   req.PapanPencatatan,
	}

	created, err := s.repo.Create(ctx, company)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	return dto.ToCompanyResponse(created), nil
}

func (s *CompanyService) Update(ctx context.Context, id int, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
	var tanggalPencatatan *time.Time
	if req.TanggalPencatatan != nil {
		t, err := time.Parse("2006-01-02", *req.TanggalPencatatan)
		if err != nil {
			return nil, fmt.Errorf("invalid tanggal_pencatatan format, use YYYY-MM-DD")
		}
		tanggalPencatatan = &t
	}

	company := &model.Company{
		Kode:              *req.Kode,
		NamaPerusahaan:    *req.NamaPerusahaan,
		TanggalPencatatan: tanggalPencatatan,
		JumlahSaham:       req.JumlahSaham,
		PapanPencatatan:   req.PapanPencatatan,
	}

	updated, err := s.repo.Update(ctx, id, company)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	return dto.ToCompanyResponse(updated), nil
}

func (s *CompanyService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *CompanyService) Import(ctx context.Context, requests []dto.CompanyImportRequest) *dto.ImportResult {
	result := &dto.ImportResult{
		TotalRows:     len(requests),
		SuccessCount:  0,
		FailureCount:  0,
		FailedRecords: []dto.FailedRecord{},
	}

	var companies []model.Company

	// Convert import requests to models
	for i, req := range requests {
		var tanggalPencatatan *time.Time
		if req.TanggalPencatatan != nil {
			t, err := time.Parse("2006-01-02", *req.TanggalPencatatan)
			if err != nil {
				result.FailedRecords = append(result.FailedRecords, dto.FailedRecord{
					RowNumber: i + 2, // +2 because CSV starts at 1 and header is skipped
					Kode:      req.Kode,
					Reason:    fmt.Sprintf("Invalid date format: %s", *req.TanggalPencatatan),
				})
				result.FailureCount++
				continue
			}
			tanggalPencatatan = &t
		}

		companies = append(companies, model.Company{
			Kode:              req.Kode,
			NamaPerusahaan:    req.NamaPerusahaan,
			TanggalPencatatan: tanggalPencatatan,
			JumlahSaham:       req.JumlahSaham,
			PapanPencatatan:   req.PapanPencatatan,
		})
	}

	// Bulk create
	created, errors, err := s.repo.BulkCreate(ctx, companies)
	if err != nil {
		s.log.Error(err)
	}

	result.SuccessCount = len(created)

	// Process errors
	for i, errMsg := range errors {
		if errMsg != nil {
			result.FailureCount++
			if i < len(requests) {
				result.FailedRecords = append(result.FailedRecords, dto.FailedRecord{
					RowNumber: i + 2,
					Kode:      requests[i].Kode,
					Reason:    errMsg.Error(),
				})
			}
		}
	}

	return result
}
