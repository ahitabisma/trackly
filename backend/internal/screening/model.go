package screening

import (
	"time"
)

type DailyScreeningResult struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ScanDate    string    `gorm:"type:date;not null;uniqueIndex:idx_screening_ticker_date" json:"scan_date"`
	Ticker      string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_screening_ticker_date" json:"ticker"`
	Rank        int       `gorm:"type:smallint;not null" json:"rank"`
	Score       float64   `gorm:"type:numeric(5,2);not null" json:"score"`
	Confidence  string    `gorm:"type:varchar(10);not null;default:low" json:"confidence"`
	Overall     string    `gorm:"type:varchar(10);not null;default:neutral" json:"overall"`
	AvgVolume   *float64  `gorm:"type:numeric(20,0)" json:"avg_volume"`
	TradingPlan *string   `gorm:"type:jsonb" json:"trading_plan"`
	AIInsight   *string   `gorm:"type:text" json:"ai_insight"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
}

func (DailyScreeningResult) TableName() string {
	return "daily_screening_results"
}
