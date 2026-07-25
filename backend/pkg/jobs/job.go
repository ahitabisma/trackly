package jobs

import (
	"context"
	"encoding/json"
	"time"
)

// JobType defines the type of job
type JobType string

const (
	JobTypeShareholdingImport JobType = "shareholding:import"
	JobTypeCompanyImport      JobType = "company:import"
	JobTypeInvestorImport     JobType = "investor:import"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusStarted   JobStatus = "started"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// Job represents a background job
type Job struct {
	ID        string          `json:"id"`
	Type      JobType         `json:"type"`
	Status    JobStatus       `json:"status"`
	Payload   json.RawMessage `json:"payload"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	StartedAt *time.Time      `json:"started_at,omitempty"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
	Progress  *Progress       `json:"progress,omitempty"`
}

// Progress tracks job progress
type Progress struct {
	Total     int    `json:"total"`
	Processed int    `json:"processed"`
	Failed    int    `json:"failed"`
	Message   string `json:"message,omitempty"`
}

// JobHandler is a function that handles job execution
type JobHandler func(ctx context.Context, job *Job) error

// JobProcessor handles job execution and updates
type JobProcessor interface {
	Process(ctx context.Context, job *Job) error
}
