package shareholding

import (
	"context"
	"encoding/json"
	"fmt"
	"trackly-backend/pkg/jobs"

	"github.com/sirupsen/logrus"
)

// ShareholdingImportPayload represents the payload for shareholding import job
type ShareholdingImportPayload struct {
	FileData []byte                  `json:"file_data"`
	FileName string                  `json:"file_name"`
	UserID   uint                    `json:"user_id"`
	Rows     []ShareHoldingImportRow `json:"rows"`
}

// NewShareholdingImportJobHandler creates a job handler for shareholding imports
func NewShareholdingImportJobHandler(
	svc *ShareHoldingService,
	logger *logrus.Logger,
) jobs.JobHandler {
	return func(ctx context.Context, job *jobs.Job) error {
		var payload ShareholdingImportPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		// Execute import
		result, appErr := svc.Import(ctx, payload.Rows, 500)
		if appErr != nil {
			return fmt.Errorf("import failed: %s", appErr.Detail)
		}

		// TODO: Store result somewhere or notify user
		logger.WithFields(map[string]interface{}{
			"job_id":        job.ID,
			"total_rows":    result.TotalRows,
			"inserted":      result.Inserted,
			"updated":       result.Updated,
			"skipped":       result.SkippedInvalid,
			"new_investors": result.NewInvestors,
			"new_companies": result.NewCompanies,
			"errors":        len(result.Errors),
		}).Info("shareholding import completed")

		return nil
	}
}
