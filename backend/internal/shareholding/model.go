package shareholding

import (
	"time"
	"trackly-backend/internal/company"
	"trackly-backend/internal/investor"
)

type Shareholding struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement"`
	CompanyID          uint      `gorm:"not null;index:idx_shareholdings_company_date;index:idx_shareholdings_company_pct"`
	InvestorID         uint      `gorm:"not null;index:idx_shareholdings_investor"`
	Date               time.Time `gorm:"type:date;not null;index:idx_shareholdings_company_date"`
	HoldingsScripless  int64     `gorm:"type:bigint;default:0"`
	HoldingsScrip      int64     `gorm:"type:bigint;default:0"`
	TotalHoldingShares int64     `gorm:"type:bigint;not null"`
	Percentage         float64   `gorm:"type:numeric(6,2);not null;index:idx_shareholdings_company_pct,sort:desc"`
	Source             *string   `gorm:"type:varchar(50)"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`

	Company  *company.Company   `gorm:"foreignKey:CompanyID"`
	Investor *investor.Investor `gorm:"foreignKey:InvestorID"`
}
