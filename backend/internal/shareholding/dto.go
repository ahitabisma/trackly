package shareholding

import "time"

type ShareHoldingImportRow struct {
	Date               time.Time
	ShareCode          string
	IssuerName         string
	InvestorName       string
	InvestorType       string
	LocalForeign       string
	Nationality        string
	Domicile           string
	HoldingScripless   int64
	HoldingScrip       int64
	TotalHoldingShares int64
	Percentage         float64
	Source             string
}

type ShareholdingImportResult struct {
	TotalRows      int
	Inserted       int
	Updated        int
	SkippedInvalid int
	NewInvestors   int
	NewCompanies   int
	Errors         []ShareholdingRowError
}

type ShareholdingRowError struct {
	RowIndex int
	Row      ShareHoldingImportRow
	Err      string
}

type ShareholdingResponse struct {
	ID                 uint    `json:"id"`
	CompanyID          uint    `json:"company_id"`
	CompanyKode        string  `json:"company_kode"`
	CompanyName        string  `json:"company_name"`
	InvestorID         uint    `json:"investor_id"`
	InvestorName       string  `json:"investor_name"`
	InvestorType       string  `json:"investor_type"`
	LocalForeign       string  `json:"local_foreign"`
	Nationality        string  `json:"nationality"`
	Domicile           string  `json:"domicile"`
	HoldingsScripless  int64   `json:"holdings_scripless"`
	HoldingsScrip      int64   `json:"holdings_scrip"`
	TotalHoldingShares int64   `json:"total_holding_shares"`
	Percentage         float64 `json:"percentage"`
	Date               string  `json:"date"`
	Source             *string `json:"source,omitempty"`
}

func ToShareholdingResponse(sh Shareholding) ShareholdingResponse {
	response := ShareholdingResponse{
		ID:                 sh.ID,
		CompanyID:          sh.CompanyID,
		InvestorID:         sh.InvestorID,
		HoldingsScripless:  sh.HoldingsScripless,
		HoldingsScrip:      sh.HoldingsScrip,
		TotalHoldingShares: sh.TotalHoldingShares,
		Percentage:         sh.Percentage,
		Date:               sh.Date.Format("2006-01-02"),
		Source:             sh.Source,
	}

	if sh.Company != nil {
		response.CompanyKode = sh.Company.Kode
		response.CompanyName = sh.Company.NamaPerusahaan
	}

	if sh.Investor != nil {
		response.InvestorName = sh.Investor.CanonicalName
		if sh.Investor.InvestorType != nil {
			response.InvestorType = *sh.Investor.InvestorType
		}
		if sh.Investor.LocalForeign != nil {
			response.LocalForeign = *sh.Investor.LocalForeign
		}
		if sh.Investor.Nationality != nil {
			response.Nationality = *sh.Investor.Nationality
		}
		if sh.Investor.Domicile != nil {
			response.Domicile = *sh.Investor.Domicile
		}
	}

	return response
}

func ToShareholdingResponseList(shareholdings []Shareholding) []ShareholdingResponse {
	if shareholdings == nil {
		return nil
	}

	responses := make([]ShareholdingResponse, len(shareholdings))
	for i, sh := range shareholdings {
		responses[i] = ToShareholdingResponse(sh)
	}
	return responses
}
