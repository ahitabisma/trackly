package company

import (
	"context"
	"encoding/json"
	"fmt"
	"trackly-backend/pkg/jobs"

	"github.com/sirupsen/logrus"
)

// CompanyImportPayload represents the payload for company import job
type CompanyImportPayload struct {
	FileData string                 `json:"file_data"`
	FileName string                 `json:"file_name"`
	Requests []CompanyImportRequest `json:"requests"`
}

// NewCompanyImportJobHandler creates a job handler for company imports
// This is an example of how to add new job types to the system
func NewCompanyImportJobHandler(
	svc *CompanyService,
	logger *logrus.Logger,
) jobs.JobHandler {
	return func(ctx context.Context, job *jobs.Job) error {
		var payload CompanyImportPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		// Execute import
		result := svc.Import(ctx, payload.Requests)

		// TODO: Store result somewhere or notify user
		logger.WithFields(map[string]interface{}{
			"job_id":     job.ID,
			"total_rows": result.TotalRows,
			"success":    result.SuccessCount,
			"failed":     result.FailureCount,
		}).Info("company import completed")

		return nil
	}
}
