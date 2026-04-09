package shareholding

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ShareHoldingRepository interface {
	UpsertBatch(ctx context.Context, rows []Shareholding) error
	FindByCompanyAndDate(ctx context.Context, companyID uint, date string) ([]Shareholding, error)
	FindByInvestor(ctx context.Context, investorID uint) ([]Shareholding, error)
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
