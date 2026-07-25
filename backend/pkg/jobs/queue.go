package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Queue defines the interface for a job queue
type Queue interface {
	Enqueue(ctx context.Context, jobType JobType, payload interface{}) (string, error)
	Dequeue(ctx context.Context, timeout time.Duration) (*Job, error)
	UpdateStatus(ctx context.Context, jobID string, status JobStatus, progress *Progress) error
	UpdateResult(ctx context.Context, jobID string, result interface{}, err string) error
	GetJob(ctx context.Context, jobID string) (*Job, error)
	GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]*Job, error)
}

// RedisQueue implements Queue using Redis
type RedisQueue struct {
	client *redis.Client
	queue  string
	status string // hash key for job status
}

// NewRedisQueue creates a new Redis-based job queue
func NewRedisQueue(redisClient *redis.Client) Queue {
	return &RedisQueue{
		client: redisClient,
		queue:  "jobs:queue",
		status: "jobs:status",
	}
}

// Enqueue adds a job to the queue
func (rq *RedisQueue) Enqueue(ctx context.Context, jobType JobType, payload interface{}) (string, error) {
	jobID := uuid.New().String()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	job := &Job{
		ID:        jobID,
		Type:      jobType,
		Status:    JobStatusPending,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("marshal job: %w", err)
	}

	// Push to queue
	pipe := rq.client.Pipeline()
	pipe.RPush(ctx, rq.queue, jobBytes)
	pipe.HSet(ctx, rq.status, jobID, string(jobBytes))
	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("enqueue job: %w", err)
	}

	return jobID, nil
}

// Dequeue retrieves a job from the queue
func (rq *RedisQueue) Dequeue(ctx context.Context, timeout time.Duration) (*Job, error) {
	// Use BLPOP with timeout (in seconds)
	timeoutSec := int64(timeout.Seconds())
	if timeoutSec < 1 {
		timeoutSec = 1
	}

	result, err := rq.client.BLPop(ctx, time.Duration(timeoutSec)*time.Second, rq.queue).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No job in queue
		}
		return nil, fmt.Errorf("dequeue: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid dequeue result")
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}

	return &job, nil
}

// UpdateStatus updates job status and progress
func (rq *RedisQueue) UpdateStatus(ctx context.Context, jobID string, status JobStatus, progress *Progress) error {
	// Get current job
	jobJson, err := rq.client.HGet(ctx, rq.status, jobID).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("job not found: %s", jobID)
		}
		return fmt.Errorf("get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobJson), &job); err != nil {
		return fmt.Errorf("unmarshal job: %w", err)
	}

	// Update status
	job.Status = status
	job.Progress = progress

	if status == JobStatusStarted && job.StartedAt == nil {
		now := time.Now()
		job.StartedAt = &now
	}

	if status == JobStatusCompleted || status == JobStatusFailed {
		now := time.Now()
		job.EndedAt = &now
	}

	jobBytes, err := json.Marshal(&job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	if err := rq.client.HSet(ctx, rq.status, jobID, string(jobBytes)).Err(); err != nil {
		return fmt.Errorf("update job status: %w", err)
	}

	return nil
}

// UpdateResult updates job result and completion status
func (rq *RedisQueue) UpdateResult(ctx context.Context, jobID string, result interface{}, errStr string) error {
	jobJson, err := rq.client.HGet(ctx, rq.status, jobID).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("job not found: %s", jobID)
		}
		return fmt.Errorf("get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobJson), &job); err != nil {
		return fmt.Errorf("unmarshal job: %w", err)
	}

	// Set result
	if result != nil {
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
		job.Result = resultBytes
	}

	// Set error
	job.Error = errStr

	// Set status
	if errStr != "" {
		job.Status = JobStatusFailed
	} else {
		job.Status = JobStatusCompleted
	}

	now := time.Now()
	job.EndedAt = &now

	jobBytes, err := json.Marshal(&job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	if err := rq.client.HSet(ctx, rq.status, jobID, string(jobBytes)).Err(); err != nil {
		return fmt.Errorf("update job result: %w", err)
	}

	return nil
}

// GetJob retrieves a specific job
func (rq *RedisQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	jobJson, err := rq.client.HGet(ctx, rq.status, jobID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobJson), &job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}

	return &job, nil
}

// GetJobsByStatus retrieves jobs with a specific status
func (rq *RedisQueue) GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]*Job, error) {
	var jobs []*Job

	// Scan all jobs in status hash
	var cursor uint64
	var count int64 = 100

	for {
		keys, nextCursor, err := rq.client.HScan(ctx, rq.status, cursor, "", count).Result()
		if err != nil {
			return nil, fmt.Errorf("scan jobs: %w", err)
		}

		for i := 0; i < len(keys); i += 2 {
			if limit > 0 && len(jobs) >= limit {
				break
			}

			jobJson := keys[i+1]
			var job Job
			if err := json.Unmarshal([]byte(jobJson), &job); err != nil {
				continue
			}

			if job.Status == status {
				jobs = append(jobs, &job)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return jobs, nil
}
