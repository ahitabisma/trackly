package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"trackly-backend/internal/dto"
)

// ParseCompanyCSV parses CSV data and returns a slice of CompanyImportRequest
func ParseCompanyCSV(reader io.Reader) ([]dto.CompanyImportRequest, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	var records []dto.CompanyImportRequest

	// Skip header row
	_, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
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

		if len(row) < 2 {
			rowNum++
			continue // Skip empty or incomplete rows
		}

		// Parse columns (skip "No" column which is index 0)
		kode := strings.TrimSpace(row[1])
		namaPerusahaan := strings.TrimSpace(row[2])

		var tanggalPencatatan *string
		if len(row) > 3 && strings.TrimSpace(row[3]) != "" {
			dateStr := parseDate(strings.TrimSpace(row[3]))
			if dateStr != "" {
				tanggalPencatatan = &dateStr
			}
		}

		var jumlahSaham *int64
		if len(row) > 4 && strings.TrimSpace(row[4]) != "" {
			if saham, err := parseNumber(strings.TrimSpace(row[4])); err == nil {
				jumlahSaham = &saham
			}
		}

		var papanPencatatan *string
		if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
			p := strings.TrimSpace(row[5])
			papanPencatatan = &p
		}

		records = append(records, dto.CompanyImportRequest{
			Kode:              kode,
			NamaPerusahaan:    namaPerusahaan,
			TanggalPencatatan: tanggalPencatatan,
			JumlahSaham:       jumlahSaham,
			PapanPencatatan:   papanPencatatan,
		})

		rowNum++
	}

	return records, nil
}

// parseDate converts date formats like "09 Des 1997" to "2006-01-02"
func parseDate(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	// Map Indonesian month names to numbers
	monthMap := map[string]string{
		"Jan": "01", "Januari": "01",
		"Feb": "02", "Februari": "02",
		"Mar": "03", "Maret": "03",
		"Apr": "04", "April": "04",
		"Mei": "05", "May": "05",
		"Jun": "06", "Juni": "06",
		"Jul": "07", "Juli": "07",
		"Agu": "08", "Agustus": "08", "Aug": "08",
		"Sep": "09", "September": "09",
		"Okt": "10", "Oktober": "10", "Oct": "10",
		"Nov": "11", "November": "11",
		"Des": "12", "Desember": "12", "Dec": "12",
	}

	parts := strings.Fields(dateStr)
	if len(parts) != 3 {
		return ""
	}

	day := parts[0]
	month := parts[1]
	year := parts[2]

	monthNum, exists := monthMap[month]
	if !exists {
		return ""
	}

	// Validate day
	dayNum, err := strconv.Atoi(day)
	if err != nil || dayNum < 1 || dayNum > 31 {
		return ""
	}

	// Format: YYYY-MM-DD
	return fmt.Sprintf("%s-%s-%s", year, monthNum, fmt.Sprintf("%02d", dayNum))
}

// parseNumber converts numbers with dot separators to int64
// e.g., "1.924.688.333" -> 1924688333
func parseNumber(numStr string) (int64, error) {
	numStr = strings.TrimSpace(numStr)
	// Remove dots (thousands separator)
	numStr = strings.ReplaceAll(numStr, ".", "")
	// Convert to int64
	return strconv.ParseInt(numStr, 10, 64)
}
