package shareholding

import (
	"context"
	"fmt"
	"trackly-backend/internal/company"
	"trackly-backend/internal/investor"
	"trackly-backend/pkg/httpx"

	"github.com/sirupsen/logrus"
)

type ShareHoldingService struct {
	companyRepo      company.CompanyRepository
	investorRepo     investor.InvestorRepository
	shareHoldingRepo ShareHoldingRepository
	logger           *logrus.Logger
}

func NewShareHoldingService(
	companyRepo company.CompanyRepository,
	investorRepo investor.InvestorRepository,
	shareHoldingRepo ShareHoldingRepository,
	logger *logrus.Logger,
) *ShareHoldingService {
	return &ShareHoldingService{
		companyRepo:      companyRepo,
		investorRepo:     investorRepo,
		shareHoldingRepo: shareHoldingRepo,
		logger:           logger,
	}
}

func (s *ShareHoldingService) Import(ctx context.Context, rows []ShareHoldingImportRow, batchSize int) (*ShareholdingImportResult, *httpx.AppError) {
	if batchSize <= 0 {
		batchSize = 500
	}

	result := &ShareholdingImportResult{
		TotalRows: len(rows),
	}

	// Cache in memory
	companyCache := make(map[string]uint)
	investorCache := make(map[string]uint)

	var pendingShareholdings []Shareholding

	flush := func() *httpx.AppError {
		if len(pendingShareholdings) == 0 {
			return nil
		}

		deduplicated := deduplicateShareholdings(pendingShareholdings)

		s.logger.WithFields(map[string]interface{}{
			"original_count":   len(pendingShareholdings),
			"deduplicated":     len(deduplicated),
			"duplicates_found": len(pendingShareholdings) - len(deduplicated),
		}).Info("shareholding batch before upsert")

		if err := s.shareHoldingRepo.UpsertBatch(ctx, deduplicated); err != nil {
			s.logger.WithError(err).WithFields(map[string]interface{}{
				"batch_size": len(deduplicated),
				"first_record": map[string]interface{}{
					"company_id":  deduplicated[0].CompanyID,
					"investor_id": deduplicated[0].InvestorID,
					"date":        deduplicated[0].Date,
				},
			}).Error("failed to upsert shareholdings batch")
			return &httpx.AppError{
				Code:   httpx.ErrInternal,
				Detail: "Failed to upsert shareholdings: " + err.Error(),
			}
		}

		result.Inserted += len(deduplicated)
		pendingShareholdings = pendingShareholdings[:0]
		return nil
	}

	for i, row := range rows {
		if err := ValidateShareholdingImportRow(row); err != nil {
			result.SkippedInvalid++
			result.Errors = append(result.Errors, ShareholdingRowError{
				RowIndex: i + 1,
				Row:      row,
				Err:      err.Error(),
			})
			continue
		}

		// Resolve company
		companyID, err := s.resolveCompany(ctx, row, companyCache, result)
		if err != nil {
			result.SkippedInvalid++
			s.logger.WithError(err).WithField("share_code", row.ShareCode).Error("failed to resolve company")
			result.Errors = append(result.Errors, ShareholdingRowError{
				RowIndex: i + 1,
				Row:      row,
				Err:      "failed to resolve company: " + err.Error(),
			})
			continue
		}

		// Resolve investor
		investorID, err := s.resolveInvestor(ctx, row, investorCache, result)
		if err != nil {
			result.SkippedInvalid++
			s.logger.WithError(err).WithField("investor_name", row.InvestorName).Error("failed to resolve investor")
			result.Errors = append(result.Errors, ShareholdingRowError{
				RowIndex: i + 1,
				Row:      row,
				Err:      "failed to resolve investor: " + err.Error(),
			})
			continue
		}

		// Validate resolved IDs
		if companyID == 0 || investorID == 0 {
			result.SkippedInvalid++
			result.Errors = append(result.Errors, ShareholdingRowError{
				RowIndex: i + 1,
				Row:      row,
				Err:      "company or investor ID is invalid",
			})
			continue
		}

		// Create shareholding record
		shareholding := Shareholding{
			CompanyID:          companyID,
			InvestorID:         investorID,
			Date:               row.Date,
			HoldingsScripless:  row.HoldingScripless,
			HoldingsScrip:      row.HoldingScrip,
			TotalHoldingShares: row.TotalHoldingShares,
			Percentage:         row.Percentage,
			Source:             &row.Source,
		}

		pendingShareholdings = append(pendingShareholdings, shareholding)

		// Flush batch if needed
		if len(pendingShareholdings) >= batchSize {
			if appErr := flush(); appErr != nil {
				return result, appErr
			}
		}
	}

	// Flush remaining
	if appErr := flush(); appErr != nil {
		return result, appErr
	}

	return result, nil
}

// deduplicateShareholdings removes duplicate shareholdings keeping the last occurrence
// Duplicates are identified by (company_id, investor_id, date) combination
func deduplicateShareholdings(shares []Shareholding) []Shareholding {
	// Key: "company_id:investor_id:date"
	seen := make(map[string]int) // maps key to index in result
	var result []Shareholding

	for _, sh := range shares {
		key := getShareholdingKey(sh)

		if idx, exists := seen[key]; exists {
			// Replace the previous occurrence with this one (keep latest)
			result[idx] = sh
		} else {
			// New combination, add to result
			seen[key] = len(result)
			result = append(result, sh)
		}
	}

	return result
}

// getShareholdingKey generates a unique key for a shareholding record
func getShareholdingKey(sh Shareholding) string {
	return sh.Date.Format("2006-01-02") + ":" +
		fmt.Sprintf("%d", sh.CompanyID) + ":" +
		fmt.Sprintf("%d", sh.InvestorID)
}
