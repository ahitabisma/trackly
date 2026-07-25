package company

import (
	"context"
	"strings"
	"trackly-backend/pkg/filter"

	"gorm.io/gorm"
)

type CompanyRepository interface {
	FindAll(ctx context.Context, fq filter.FilteringQuery) ([]*Company, int64, error)
	FindByID(ctx context.Context, id int) (*Company, error)
	Create(ctx context.Context, company *Company) (*Company, error)
	Update(ctx context.Context, id int, req *UpdateCompanyRequest) (*Company, error)
	Delete(ctx context.Context, id int) error
	BulkCreate(ctx context.Context, companies []Company) ([]Company, error)
	FindOrCreateByKode(ctx context.Context, company *Company) (*Company, error)
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) FindAll(ctx context.Context, fq filter.FilteringQuery) ([]*Company, int64, error) {
	allowed := []string{"kode", "nama_perusahaan"}
	db := r.db.WithContext(ctx).Model(&Company{})

	// Apply filters
	db = filter.ApplyGormFilter(db, fq, allowed)

	// Count total records before pagination
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Default ordering
	order := "kode ASC"
	if fq.OrderKey != "" {
		rule := "DESC"
		if fq.OrderRule == "asc" {
			rule = "ASC"
		}
		order = fq.OrderKey + " " + rule
	}
	db = db.Order(order)

	// Pagination
	db, _, _ = filter.ApplyPagination(db, fq)

	var result []*Company
	if err := db.Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (r *companyRepository) FindByID(ctx context.Context, id int) (*Company, error) {
	var company Company
	if err := r.db.WithContext(ctx).First(&company, id).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Create(ctx context.Context, company *Company) (*Company, error) {
	if err := r.db.WithContext(ctx).Create(company).Error; err != nil {
		if strings.Contains(err.Error(), `unique constraint "companies_kode_key"`) {
			return nil, gorm.ErrDuplicatedKey
		}
		return nil, err
	}
	return company, nil
}

func (r *companyRepository) Update(ctx context.Context, id int, req *UpdateCompanyRequest) (*Company, error) {
	var company Company

	if err := r.db.WithContext(ctx).First(&company, id).Error; err != nil {
		return nil, err
	}

	if req.Kode != nil {
		company.Kode = *req.Kode
	}
	if req.NamaPerusahaan != nil {
		company.NamaPerusahaan = *req.NamaPerusahaan
	}
	if req.TanggalPencatatan != nil {
		parsedTime := ParseDatePointer(req.TanggalPencatatan)
		if parsedTime == nil {
			return nil, gorm.ErrInvalidData
		}
		company.TanggalPencatatan = parsedTime
	}

	if req.JumlahSaham != nil {
		company.JumlahSaham = req.JumlahSaham
	}
	if req.PapanPencatatan != nil {
		company.PapanPencatatan = req.PapanPencatatan
	}

	if err := r.db.WithContext(ctx).Save(&company).Error; err != nil {
		if strings.Contains(err.Error(), `unique constraint "companies_kode_key"`) {
			return nil, gorm.ErrDuplicatedKey
		}

		return nil, err
	}

	return &company, nil
}

func (r *companyRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&Company{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *companyRepository) BulkCreate(ctx context.Context, companies []Company) ([]Company, error) {
	tx := r.db.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	if err := tx.Create(&companies).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return companies, nil
}

func (r *companyRepository) FindOrCreateByKode(ctx context.Context, company *Company) (*Company, error) {
	result := r.db.WithContext(ctx).
		Where(Company{Kode: company.Kode}).
		Attrs(company).
		FirstOrCreate(company)
	return company, result.Error
}
