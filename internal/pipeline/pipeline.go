package pipeline

import (
	"context"
	"sync"
)

type Pipeline[J any] struct {
	pool    *WorkerPool[J]
	fetchFn func(ctx context.Context) ([]J, error)
	// TODO: add logger
}

func NewPipeline[J any](
	workerCount int,
	handler Handler[J],
	fetchFn func(ctx context.Context) ([]J, error),
) *Pipeline[J] {
	return &Pipeline[J]{
		pool:    NewWorkerPool(workerCount, handler),
		fetchFn: fetchFn,
	}
}

// Run fetches all jobs via fetchFn and dispatches to the worker pool.
func (p *Pipeline[J]) Run(ctx context.Context) error {
	jobs, err := p.fetchFn(ctx)
	if err != nil {
		return err
	}
	return p.run(ctx, jobs)
}

// RunForUser dispatches a single job — fetchFn is bypassed.
func (p *Pipeline[J]) RunForOne(ctx context.Context, job J) error {
	return p.run(ctx, []J{job})
}

// run is the shared dispatch path.
func (p *Pipeline[J]) run(ctx context.Context, jobs []J) error {
	var wg sync.WaitGroup

	p.pool.Start(ctx, &wg)
	p.pool.Dispatch(ctx, jobs)

	wg.Wait()
	close(p.pool.results)
	p.pool.Wait()

	// TODO: return aggregated error
	return nil
}
