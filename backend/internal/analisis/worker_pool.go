package analisis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var ErrWorkerPoolBusy = fmt.Errorf("all workers busy, try again later")

const (
	poolSize       = 10
	acquireTimeout = 30 * time.Second
	watchdogLimit  = 2 * time.Minute
)

type WorkerPool struct {
	mu      sync.Mutex
	workers [poolSize]bool
	sem     chan struct{}
	log     *logrus.Logger
}

func NewWorkerPool(log *logrus.Logger) *WorkerPool {
	p := &WorkerPool{
		sem: make(chan struct{}, poolSize),
		log: log,
	}
	for i := 0; i < poolSize; i++ {
		p.sem <- struct{}{}
	}
	return p
}

func (p *WorkerPool) AcquireWorker(ctx context.Context) (int, error) {
	select {
	case <-p.sem:
	case <-time.After(acquireTimeout):
		return 0, ErrWorkerPoolBusy
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for i, busy := range p.workers {
		if !busy {
			p.workers[i] = true
			slot := i + 1
			p.log.WithField("slot", slot).Info("worker acquired")
			return slot, nil
		}
	}
	return 0, fmt.Errorf("worker pool: no idle worker despite semaphore")
}

func (p *WorkerPool) ReleaseWorker(n int) {
	if n < 1 || n > poolSize {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.workers[n-1] {
		return
	}
	p.workers[n-1] = false
	select {
	case p.sem <- struct{}{}:
	default:
	}
	p.log.WithField("slot", n).Info("worker released")
}

func (p *WorkerPool) Watchdog(n int, startedAt time.Time) {
	go func() {
		deadline := watchdogLimit - time.Since(startedAt)
		if deadline <= 0 {
			deadline = time.Second
		}
		<-time.After(deadline)
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.workers[n-1] {
			p.log.WithField("slot", n).Warn("worker stuck, force-releasing")
			p.workers[n-1] = false
			select {
			case p.sem <- struct{}{}:
			default:
			}
		}
	}()
}
