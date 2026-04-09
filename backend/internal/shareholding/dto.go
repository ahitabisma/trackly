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
