package model

import "time"

type Shareholding struct {
	ID         int       `db:"id"`
	ShareCode  string    `db:"share_code"`
	IssuerName string    `db:"issuer_name"`
	ReportDate time.Time `db:"report_date"`

	InvestorName string  `db:"investor_name"`
	InvestorType string  `db:"investor_type"`
	LocalForeign string  `db:"local_foreign"`
	Nationality  *string `db:"nationality"`
	Domicile     *string `db:"domicile"`

	HoldingsScripless  int64   `db:"holdings_scripless"`
	HoldingsScrip      int64   `db:"holdings_scrip"`
	TotalHoldingShares int64   `db:"total_holding_shares"`
	Percentage         float64 `db:"percentage"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
