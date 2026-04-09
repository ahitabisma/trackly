package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Worker processes jobs from the queue
type Worker struct {
	queue       Queue
	handlers    map[JobType]JobHandler
	logger      *logrus.Logger
	workers     int
	stopChan    chan struct{}
	wg          sync.WaitGroup
	handlerLock sync.RWMutex
}

// NewWorker creates a new job worker
func NewWorker(queue Queue, numWorkers int, logger *logrus.Logger) *Worker {
	if numWorkers < 1 {
		numWorkers = 2
	}
	return &Worker{
		queue:    queue,
		handlers: make(map[JobType]JobHandler),
		logger:   logger,
		workers:  numWorkers,
		stopChan: make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a job type
func (w *Worker) RegisterHandler(jobType JobType, handler JobHandler) {
	w.handlerLock.Lock()
	defer w.handlerLock.Unlock()
	w.handlers[jobType] = handler
}

// Start starts the worker goroutines
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.processJobs(ctx)
	}
	w.logger.Infof("started %d job workers", w.workers)
}

// Stop stops all worker goroutines
func (w *Worker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	w.logger.Info("job workers stopped")
}

// processJobs processes jobs from the queue continuously
func (w *Worker) processJobs(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopChan:
			return
		default:
		}

		// Dequeue with 5 second timeout
		job, err := w.queue.Dequeue(ctx, 5*time.Second)
		if err != nil {
			w.logger.WithError(err).Error("failed to dequeue job")
			continue
		}

		// No job available
		if job == nil {
			continue
		}

		w.executeJob(ctx, job)
	}
}

// executeJob executes a single job
func (w *Worker) executeJob(ctx context.Context, job *Job) {
	w.handlerLock.RLock()
	handler, exists := w.handlers[job.Type]
	w.handlerLock.RUnlock()

	if !exists {
		errMsg := fmt.Sprintf("no handler for job type: %s", job.Type)
		w.logger.WithField("job_id", job.ID).Error(errMsg)
		if err := w.queue.UpdateResult(ctx, job.ID, nil, errMsg); err != nil {
			w.logger.WithError(err).Error("failed to update job result")
		}
		return
	}

	// Update status to started
	if err := w.queue.UpdateStatus(ctx, job.ID, JobStatusStarted, nil); err != nil {
		w.logger.WithError(err).Error("failed to update job status to started")
	}

	w.logger.WithFields(map[string]interface{}{
		"job_id":   job.ID,
		"job_type": job.Type,
	}).Info("processing job")

	// Execute handler
	err := handler(ctx, job)

	// Update status to completed or failed
	if err != nil {
		w.logger.WithFields(map[string]interface{}{
			"job_id":   job.ID,
			"job_type": job.Type,
			"error":    err.Error(),
		}).Error("job execution failed")
		if updateErr := w.queue.UpdateResult(ctx, job.ID, nil, err.Error()); updateErr != nil {
			w.logger.WithError(updateErr).Error("failed to update job result")
		}
	} else {
		w.logger.WithFields(map[string]interface{}{
			"job_id":   job.ID,
			"job_type": job.Type,
		}).Info("job completed successfully")
		if updateErr := w.queue.UpdateResult(ctx, job.ID, map[string]bool{"success": true}, ""); updateErr != nil {
			w.logger.WithError(updateErr).Error("failed to update job result")
		}
	}
}
