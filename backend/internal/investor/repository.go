package investor

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InvestorRepository interface {
	FindByNormalized(ctx context.Context, normalizedName string) (*Investor, error)
	FindAliasByNormalized(ctx context.Context, normalizedAlias string) (*InvestorAlias, error)
	Create(ctx context.Context, investor *Investor) error
	CreateAlias(ctx context.Context, alias *InvestorAlias) error
	FindDuplicateCandidates(ctx context.Context, threshold float64) ([]DuplicateCandidate, error)
	MergeInvestors(ctx context.Context, keepID, mergeID uint) error
}

type investorRepository struct {
	db *gorm.DB
}

func NewInvestorRepository(db *gorm.DB) InvestorRepository {
	return &investorRepository{db: db}
}

func (r *investorRepository) FindByNormalized(ctx context.Context, normalizedName string) (*Investor, error) {
	var investor Investor
	err := r.db.WithContext(ctx).
		Where("normalized_name = ?", normalizedName).
		First(&investor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &investor, err
}

func (r *investorRepository) FindAliasByNormalized(ctx context.Context, normalizedAlias string) (*InvestorAlias, error) {
	var alias InvestorAlias
	err := r.db.WithContext(ctx).
		Where("normalized_alias = ?", normalizedAlias).
		First(&alias).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &alias, err
}

func (r *investorRepository) Create(ctx context.Context, investor *Investor) error {
	return r.db.WithContext(ctx).Create(investor).Error
}

func (r *investorRepository) CreateAlias(ctx context.Context, alias *InvestorAlias) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "normalized_alias"}},
			DoNothing: true,
		}).
		Create(alias).Error
}

func (r *investorRepository) FindDuplicateCandidates(ctx context.Context, threshold float64) ([]DuplicateCandidate, error) {
	var results []DuplicateCandidate
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			i1.id            AS investor_a_id,
			i1.canonical_name AS investor_a_name,
			i2.id            AS investor_b_id,
			i2.canonical_name AS investor_b_name,
			similarity(i1.normalized_name, i2.normalized_name) AS score
		FROM investors i1
		JOIN investors i2 ON i1.id < i2.id
		WHERE similarity(i1.normalized_name, i2.normalized_name) > ?
		ORDER BY score DESC
	`, threshold).Scan(&results).Error
	return results, err
}

func (r *investorRepository) MergeInvestors(ctx context.Context, keepID, mergeID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Pindahkan shareholdings, skip jika sudah ada snapshot yang sama
		if err := tx.Exec(`
			UPDATE shareholdings SET investor_id = ?
			WHERE investor_id = ?
			AND NOT EXISTS (
				SELECT 1 FROM shareholdings s2
				WHERE s2.investor_id = ?
				  AND s2.company_id = shareholdings.company_id
				  AND s2.date       = shareholdings.date
			)
		`, keepID, mergeID, keepID).Error; err != nil {
			return err
		}

		// Pindahkan aliases
		if err := tx.Model(&InvestorAlias{}).
			Where("investor_id = ?", mergeID).
			Update("investor_id", keepID).Error; err != nil {
			return err
		}

		// Hapus investor lama
		return tx.Delete(&Investor{}, mergeID).Error
	})
}
