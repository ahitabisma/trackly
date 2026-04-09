package shareholding

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"trackly-backend/internal/company"
	"trackly-backend/internal/investor"
)

func ValidateShareholdingImportRow(row ShareHoldingImportRow) error {
	var errs []string
	if strings.TrimSpace(row.ShareCode) == "" {
		errs = append(errs, "share_code kosong")
	}
	if strings.TrimSpace(row.InvestorName) == "" {
		errs = append(errs, "investor_name kosong")
	}
	if row.Date.IsZero() {
		errs = append(errs, "date tidak valid")
	}
	if row.Percentage < 0 || row.Percentage > 100 {
		errs = append(errs, fmt.Sprintf("percentage tidak valid: %.2f", row.Percentage))
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

func (svc *ShareHoldingService) resolveCompany(
	ctx context.Context,
	row ShareHoldingImportRow,
	cache map[string]uint,
	result *ShareholdingImportResult,
) (uint, error) {
	if id, ok := cache[row.ShareCode]; ok {
		return id, nil
	}

	company := &company.Company{
		Kode:           row.ShareCode,
		NamaPerusahaan: row.IssuerName,
	}

	created, err := svc.companyRepo.FindOrCreateByKode(ctx, company)
	if err != nil {
		return 0, err
	}

	// Deteksi apakah record baru (CreatedAt == UpdatedAt dan sangat baru)
	if isNewRecord(created.CreatedAt) {
		result.NewCompanies++
		svc.logger.Info("new company", "kode", created.Kode, "nama", created.NamaPerusahaan)
	}

	cache[row.ShareCode] = created.ID
	return created.ID, nil
}

func (svc *ShareHoldingService) resolveInvestor(
	ctx context.Context,
	row ShareHoldingImportRow,
	cache map[string]uint,
	result *ShareholdingImportResult,
) (uint, error) {
	normalized := investor.NormalizeInvestorName(row.InvestorName)

	if id, ok := cache[normalized]; ok {
		return id, nil
	}

	source := row.Source

	// 1. Cek alias
	alias, err := svc.investorRepo.FindAliasByNormalized(ctx, normalized)
	if err != nil {
		return 0, err
	}
	if alias != nil {
		cache[normalized] = alias.InvestorID
		return alias.InvestorID, nil
	}

	// 2. Cek investor langsung (mungkin masuk lewat path lain)
	existingInvestor, err := svc.investorRepo.FindByNormalized(ctx, normalized)
	if err != nil {
		return 0, err
	}
	if existingInvestor != nil {
		// Daftarkan alias baru supaya lookup berikutnya lebih cepat
		_ = svc.investorRepo.CreateAlias(ctx, &investor.InvestorAlias{
			InvestorID:      existingInvestor.ID,
			AliasName:       row.InvestorName,
			NormalizedAlias: normalized,
			Source:          &source,
		})
		cache[normalized] = existingInvestor.ID
		return existingInvestor.ID, nil
	}

	// 3. Buat investor baru
	invType := nullableString(row.InvestorType)
	lf := nullableString(row.LocalForeign)
	nationality := nullableString(row.Nationality)
	domicile := nullableString(row.Domicile)

	newInvestor := &investor.Investor{
		CanonicalName:  row.InvestorName,
		NormalizedName: normalized,
		InvestorType:   invType,
		LocalForeign:   lf,
		Nationality:    nationality,
		Domicile:       domicile,
	}
	if err := svc.investorRepo.Create(ctx, newInvestor); err != nil {
		return 0, fmt.Errorf("create investor %q: %w", row.InvestorName, err)
	}

	_ = svc.investorRepo.CreateAlias(ctx, &investor.InvestorAlias{
		InvestorID:      newInvestor.ID,
		AliasName:       row.InvestorName,
		NormalizedAlias: normalized,
		Source:          &source,
	})

	result.NewInvestors++
	svc.logger.Info("new investor",
		"canonical", newInvestor.CanonicalName,
		"normalized", normalized,
	)

	cache[normalized] = newInvestor.ID
	return newInvestor.ID, nil
}

func isNewRecord(createdAt time.Time) bool {
	return time.Since(createdAt) < 5*time.Second
}

func nullableString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// ParseShareholdingCSV parses shareholding CSV file
// Uses column headers to robustly find the correct columns
func ParseShareholdingCSV(reader io.Reader) ([]ShareHoldingImportRow, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	var records []ShareHoldingImportRow

	// Read and parse header
	headerRow, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column names to indices
	columnMap := make(map[string]int)
	for i, col := range headerRow {
		normalizedCol := strings.ToLower(strings.TrimSpace(col))
		columnMap[normalizedCol] = i
	}

	// Define required columns
	requiredCols := map[string]string{
		"date":                 "date",
		"share_code":           "shareCode",
		"issuer_name":          "issuerName",
		"investor_name":        "investorName",
		"investor_type":        "investorType",
		"local_foreign":        "localForeign",
		"nationality":          "nationality",
		"domicile":             "domicile",
		"holdings_scripless":   "holdingsScripless",
		"holdings_scrip":       "holdingsScrip",
		"total_holding_shares": "totalHoldingShares",
		"percentage":           "percentage",
	}

	// Verify all required columns exist
	for colName := range requiredCols {
		if _, exists := columnMap[colName]; !exists {
			return nil, fmt.Errorf("missing required column: %s", colName)
		}
	}

	rowNum := 2 // Start from 2 since we skipped header
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row %d: %w", rowNum, err)
		}

		// Parse columns using the column map
		dateStr := ""
		if idx, ok := columnMap["date"]; ok && idx < len(row) {
			dateStr = strings.TrimSpace(row[idx])
		}

		shareCode := ""
		if idx, ok := columnMap["share_code"]; ok && idx < len(row) {
			shareCode = strings.TrimSpace(row[idx])
		}

		issuerName := ""
		if idx, ok := columnMap["issuer_name"]; ok && idx < len(row) {
			issuerName = strings.TrimSpace(row[idx])
		}

		investorName := ""
		if idx, ok := columnMap["investor_name"]; ok && idx < len(row) {
			investorName = strings.TrimSpace(row[idx])
		}

		investorType := ""
		if idx, ok := columnMap["investor_type"]; ok && idx < len(row) {
			investorType = strings.TrimSpace(row[idx])
		}

		localForeign := ""
		if idx, ok := columnMap["local_foreign"]; ok && idx < len(row) {
			localForeign = strings.TrimSpace(row[idx])
		}

		nationality := ""
		if idx, ok := columnMap["nationality"]; ok && idx < len(row) {
			nationality = strings.TrimSpace(row[idx])
		}

		domicile := ""
		if idx, ok := columnMap["domicile"]; ok && idx < len(row) {
			domicile = strings.TrimSpace(row[idx])
		}

		holdingsScripless := int64(0)
		if idx, ok := columnMap["holdings_scripless"]; ok && idx < len(row) {
			holdingsScripless = parseNumber(strings.TrimSpace(row[idx]))
		}

		holdingsScrip := int64(0)
		if idx, ok := columnMap["holdings_scrip"]; ok && idx < len(row) {
			holdingsScrip = parseNumber(strings.TrimSpace(row[idx]))
		}

		totalHoldingShares := int64(0)
		if idx, ok := columnMap["total_holding_shares"]; ok && idx < len(row) {
			totalHoldingShares = parseNumber(strings.TrimSpace(row[idx]))
		}

		percentage := 0.0
		if idx, ok := columnMap["percentage"]; ok && idx < len(row) {
			percentage = parsePercentage(strings.TrimSpace(row[idx]))
		}

		date, err := parseShareholdingDate(dateStr)
		if err != nil {
			rowNum++
			continue // Skip rows with invalid dates
		}

		records = append(records, ShareHoldingImportRow{
			Date:               date,
			ShareCode:          shareCode,
			IssuerName:         issuerName,
			InvestorName:       investorName,
			InvestorType:       investorType,
			LocalForeign:       localForeign,
			Nationality:        nationality,
			Domicile:           domicile,
			HoldingScripless:   holdingsScripless,
			HoldingScrip:       holdingsScrip,
			TotalHoldingShares: totalHoldingShares,
			Percentage:         percentage,
			Source:             "csv_import",
		})

		rowNum++
	}

	return records, nil
}

// parseShareholdingDate parses dates like "31-Mar-2026" to time.Time
func parseShareholdingDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// Map Indonesian month abbreviations
	monthMap := map[string]int{
		"Jan": 1, "Januari": 1,
		"Feb": 2, "Februari": 2,
		"Mar": 3, "Maret": 3, "Mrt": 3,
		"Apr": 4, "April": 4,
		"Mei": 5, "May": 5,
		"Jun": 6, "Juni": 6,
		"Jul": 7, "Juli": 7,
		"Agu": 8, "Agustus": 8, "Aug": 8,
		"Sep": 9, "September": 9,
		"Okt": 10, "Oktober": 10, "Oct": 10,
		"Nov": 11, "November": 11,
		"Des": 12, "Desember": 12, "Dec": 12,
	}

	// Format: "31-Mar-2026"
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}

	day, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("invalid day: %s", parts[0])
	}

	monthStr := strings.TrimSpace(parts[1])
	month, exists := monthMap[monthStr]
	if !exists {
		return time.Time{}, fmt.Errorf("invalid month: %s", monthStr)
	}

	year, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil || year < 1900 {
		return time.Time{}, fmt.Errorf("invalid year: %s", parts[2])
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// parseNumber converts numbers with dot separators to int64
// e.g., "3.200.142.830" -> 3200142830
func parseNumber(numStr string) int64 {
	numStr = strings.TrimSpace(numStr)
	if numStr == "" || numStr == "-" {
		return 0
	}
	// Remove dots (thousands separator)
	numStr = strings.ReplaceAll(numStr, ".", "")
	val, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// parsePercentage converts percentage strings with comma decimal separator
// e.g., "41,10" -> 41.10
func parsePercentage(percStr string) float64 {
	percStr = strings.TrimSpace(percStr)
	if percStr == "" || percStr == "-" {
		return 0.0
	}
	// Replace comma with dot for decimal separator
	percStr = strings.ReplaceAll(percStr, ",", ".")
	val, err := strconv.ParseFloat(percStr, 64)
	if err != nil {
		return 0.0
	}
	return val
}
