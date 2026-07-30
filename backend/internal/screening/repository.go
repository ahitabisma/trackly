package screening

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type ScreeningResult struct {
	ScanDate    string
	Ticker      string
	Rank        int
	Score       float64
	Confidence  string
	Overall     string
	AvgVolume   float64
	TradingPlan string
	AIInsight   string
}

func (r *Repository) Upsert(ctx context.Context, row *ScreeningResult) error {
	var tradingPlan *string
	if row.TradingPlan != "" {
		tradingPlan = &row.TradingPlan
	}
	var aiInsight *string
	if row.AIInsight != "" {
		aiInsight = &row.AIInsight
	}
	var avgVol *float64
	if row.AvgVolume > 0 {
		avgVol = &row.AvgVolume
	}

	return r.db.WithContext(ctx).Exec(`
		INSERT INTO daily_screening_results (scan_date, ticker, rank, score, confidence, overall, avg_volume, trading_plan, ai_insight)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULLIF(?,'')::jsonb, ?)
		ON CONFLICT (scan_date, ticker) DO UPDATE SET
			rank = EXCLUDED.rank,
			score = EXCLUDED.score,
			confidence = EXCLUDED.confidence,
			overall = EXCLUDED.overall,
			avg_volume = EXCLUDED.avg_volume,
			trading_plan = EXCLUDED.trading_plan,
			ai_insight = EXCLUDED.ai_insight
	`, row.ScanDate, row.Ticker, row.Rank, row.Score, row.Confidence, row.Overall, avgVol, tradingPlan, aiInsight).Error
}

func (r *Repository) GetByDate(ctx context.Context, scanDate string) ([]DailyScreeningResult, error) {
	var results []DailyScreeningResult
	err := r.db.WithContext(ctx).
		Where("scan_date = ?", scanDate).
		Order("rank ASC").
		Find(&results).Error
	return results, err
}

func (r *Repository) GetLatest(ctx context.Context) ([]DailyScreeningResult, error) {
	var results []DailyScreeningResult
	err := r.db.WithContext(ctx).
		Order("scan_date DESC, rank ASC").
		Limit(10).
		Find(&results).Error
	return results, err
}
