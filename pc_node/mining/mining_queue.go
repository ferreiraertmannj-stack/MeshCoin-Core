package mining

import (
	"context"
	"fmt"
	"sync"
)

type MiningQueue struct {
	jobs    chan *MiningJob
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewMiningQueue(size int, workers int) *MiningQueue {
	return &MiningQueue{
		jobs:    make(chan *MiningJob, size),
		workers: workers,
	}
}

func (q *MiningQueue) Start(ctx context.Context) {
	q.ctx, q.cancel = context.WithCancel(ctx)
}

func (q *MiningQueue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
	q.wg.Wait()
}

func (q *MiningQueue) Enqueue(job *MiningJob) error {
	select {
	case <-q.ctx.Done():
		return fmt.Errorf("mining queue stopped")
	case q.jobs <- job:
		return nil
	default:
		return fmt.Errorf("mining queue full (backpressure)")
	}
}

func (q *MiningQueue) LaunchWorkers(
	pipeline HashPipeline,
	validator *ShareValidator,
	generator *NonceGenerator,
	events MiningEvents,
	stats *MiningStatistics,
	network NetworkProvider,
) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		worker := NewPoWWorker(i, pipeline, validator, generator, events, stats, network)

		go func() {
			defer q.wg.Done()
			for {
				select {
				case <-q.ctx.Done():
					return
				case job := <-q.jobs:
					worker.Mine(job)
				}
			}
		}()
	}
}
