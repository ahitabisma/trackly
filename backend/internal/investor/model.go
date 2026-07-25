package investor

import "time"

type Investor struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	CanonicalName  string    `gorm:"type:varchar(255);not null"`
	NormalizedName string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_investors_normalized"`
	InvestorType   *string   `gorm:"type:varchar(20)"` // ID / CP / etc
	LocalForeign   *string   `gorm:"type:char(1)"`     // L / F
	Nationality    *string   `gorm:"type:varchar(100)"`
	Domicile       *string   `gorm:"type:varchar(100)"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

type InvestorAlias struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	InvestorID      uint      `gorm:"not null;index:idx_aliases_investor"`
	AliasName       string    `gorm:"type:varchar(255);not null"`                                    // nama asli dari raw data
	NormalizedAlias string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_aliases_normalized"` // versi normalized
	Source          *string   `gorm:"type:varchar(50)"`                                              // dari file / API mana
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}
