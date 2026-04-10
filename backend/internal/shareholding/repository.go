package shareholding

import (
	"context"
	"trackly-backend/pkg/filter"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ShareHoldingRepository interface {
	UpsertBatch(ctx context.Context, rows []Shareholding) error
	FindByCompanyAndDate(ctx context.Context, companyID uint, date string) ([]Shareholding, error)
	FindByInvestor(ctx context.Context, investorID uint) ([]Shareholding, error)
	GetAll(ctx context.Context, fq filter.FilteringQuery) ([]Shareholding, int64, error)
}

type shareholdingRepository struct {
	db *gorm.DB
}

func NewShareholdingRepository(db *gorm.DB) ShareHoldingRepository {
	return &shareholdingRepository{db: db}
}

func (r *shareholdingRepository) UpsertBatch(ctx context.Context, rows []Shareholding) error {
	if len(rows) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "company_id"},
				{Name: "investor_id"},
				{Name: "date"},
			},
			UpdateAll: true,
		}).
		CreateInBatches(rows, 100).Error
}

func (r *shareholdingRepository) FindByCompanyAndDate(ctx context.Context, companyID uint, date string) ([]Shareholding, error) {
	var rows []Shareholding
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND date = ?", companyID, date).
		Order("percentage DESC").
		Preload("Investor").
		Find(&rows).Error
	return rows, err
}

func (r *shareholdingRepository) FindByInvestor(ctx context.Context, investorID uint) ([]Shareholding, error) {
	var rows []Shareholding
	err := r.db.WithContext(ctx).
		Where("investor_id = ?", investorID).
		Order("date DESC").
		Preload("Company").
		Find(&rows).Error
	return rows, err
}

func (r *shareholdingRepository) GetAll(ctx context.Context, fq filter.FilteringQuery) ([]Shareholding, int64, error) {
	allowed := []string{"id", "company_id", "investor_id", "date", "holdings_scripless", "holdings_scrip", "total_holding_shares", "percentage", "source"}
	db := r.db.WithContext(ctx).Model(&Shareholding{})

	db = db.Joins("Company").Joins("Investor")

	// Apply filters on shareholdings table
	db = filter.ApplyGormFilter(db, fq, allowed)

	// Apply filters on joined tables
	if fq.Filters != nil {
		if companyKode, ok := fq.Filters["company_kode"]; ok {
			db = db.Where(`"Company"."kode" = ?`, companyKode)
		}
		if companyName, ok := fq.Filters["company_name"]; ok {
			db = db.Where(`"Company"."nama_perusahaan" ILIKE ?`, "%"+companyName.(string)+"%")
		}
		if investorName, ok := fq.Filters["investor_name"]; ok {
			db = db.Where(`"Investor"."canonical_name" ILIKE ?`, "%"+investorName.(string)+"%")
		}
		if investorType, ok := fq.Filters["investor_type"]; ok {
			db = db.Where(`"Investor"."investor_type" = ?`, investorType)
		}
		if localForeign, ok := fq.Filters["local_foreign"]; ok {
			db = db.Where(`"Investor"."local_foreign" = ?`, localForeign)
		}
		if nationality, ok := fq.Filters["nationality"]; ok {
			db = db.Where(`"Investor"."nationality" = ?`, nationality)
		}
		if domicile, ok := fq.Filters["domicile"]; ok {
			db = db.Where(`"Investor"."domicile" = ?`, domicile)
		}
	}

	// Count total records before pagination
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Default ordering by company_kode
	order := `"Company"."kode" ASC`
	if fq.OrderKey != "" {
		rule := "ASC"
		if fq.OrderRule == "desc" {
			rule = "DESC"
		}
		switch fq.OrderKey {
		case "company_kode":
			order = `"Company"."kode" ` + rule
		case "company_name":
			order = `"Company"."nama_perusahaan" ` + rule
		case "investor_name":
			order = `"Investor"."canonical_name" ` + rule
		case "investor_type":
			order = `"Investor"."investor_type" ` + rule
		case "local_foreign":
			order = `"Investor"."local_foreign" ` + rule
		case "nationality":
			order = `"Investor"."nationality" ` + rule
		case "domicile":
			order = `"Investor"."domicile" ` + rule
		default:
			order = fq.OrderKey + " " + rule
		}
	}
	db = db.Order(order)

	// Pagination
	db, _, _ = filter.ApplyPagination(db, fq)

	var result []Shareholding
	if err := db.Preload("Company").Preload("Investor").Find(&result).Error; err != nil {
		return nil, 0, err
	}
	return result, total, nil
}
