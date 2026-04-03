package dto

import (
	"trackly-backend/internal/model"
)

type ShareholdingResponse struct {
	ShareCode    string `json:"share_code"`
	InvestorName string `json:"investor_name"`
	InvestorType string `json:"investor_type"`
	Category     string `json:"category"`
	Origin       string `json:"origin"`

	TotalShares int64   `json:"total_shares"`
	Percentage  float64 `json:"percentage"`
}

func ToShareholdingResponse(m model.Shareholding) ShareholdingResponse {
	return ShareholdingResponse{
		ShareCode:    m.ShareCode,
		InvestorName: m.InvestorName,
		InvestorType: m.InvestorType,
		Origin:       m.LocalForeign,
		TotalShares:  m.TotalHoldingShares,
		Percentage:   m.Percentage,
	}
}
