package pipeline

import (
	"context"
	"sync"
	"time"
	// TODO: import logger
)

// Handler is the function signature the worker calls for each job.
// Implemented by the service layer — injected at construction.
type Handler[J any] func(ctx context.Context, job J) error

// WorkerPool is a generic bounded pool of goroutines.
// J is the job type — one pool per pipeline.
type WorkerPool[J any] struct {
	workerCount int
	jobs        chan J
	results     chan Result[J]
	handler     Handler[J]
	collectDone chan struct{}
	// TODO: add logger
}

// NewWorkerPool constructs a WorkerPool for job type J.
func NewWorkerPool[J any](workerCount int, handler Handler[J]) *WorkerPool[J] {
	return &WorkerPool[J]{
		workerCount: workerCount,
		handler:     handler,
		collectDone: make(chan struct{}),
		jobs:        make(chan J, workerCount*2),
		results:     make(chan Result[J], workerCount),
	}
}

// Start launches N worker goroutines and one collector goroutine.
func (wp *WorkerPool[J]) Start(ctx context.Context, wg *sync.WaitGroup) {
	for i := 0; i < wp.workerCount; i++ {
		wg.Add(1)
		go wp.work(ctx, wg)
	}

	go wp.collect()
}

// work processes jobs until the channel is closed or ctx is cancelled.
func (wp *WorkerPool[J]) work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range wp.jobs {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			err := wp.handler(ctx, job)
			wp.results <- Result[J]{Job: job, Err: err, Duration: time.Since(start)}
		}
	}
}

// collect drains the results channel until it is closed.
// Signals completion via collectDone.
func (wp *WorkerPool[J]) collect() {
	defer close(wp.collectDone)

	for result := range wp.results {
		if result.succeeded() {
			// TODO: log success with job details and duration
		} else {
			// TODO: log error with job details and err
		}
		// TODO: emit metrics (success/error counters, duration histogram)
	}
}

// Dispatch pushes all jobs into the jobs channel then closes it.
// Closing signals workers to stop ranging once drained.
func (wp *WorkerPool[J]) Dispatch(ctx context.Context, jobs []J) {
	defer close(wp.jobs)

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return
		case wp.jobs <- job:
		}
	}
}

// Wait blocks until the collector has finished draining.
// Must be called after wg.Wait() and close(wp.results).
func (wp *WorkerPool[J]) Wait() {
	<-wp.collectDone
}
