package storage

import (
	"context"
	"fmt"
	"sync"
)

type StorageJob struct {
	Action string
	Data   interface{}
	Done   chan error
}

type StorageQueue struct {
	jobs    chan *StorageJob
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	engine  *StorageEngine
}

func NewStorageQueue(size int, workers int, engine *StorageEngine) *StorageQueue {
	return &StorageQueue{
		jobs:    make(chan *StorageJob, size),
		workers: workers,
		engine:  engine,
	}
}

func (q *StorageQueue) Start(ctx context.Context) {
	q.ctx, q.cancel = context.WithCancel(ctx)
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *StorageQueue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
	q.wg.Wait()
}

func (q *StorageQueue) Enqueue(job *StorageJob) error {
	select {
	case <-q.ctx.Done():
		return fmt.Errorf("storage queue stopped")
	case q.jobs <- job:
		return <-job.Done // Synchronous wait for consistency in this mock
	default:
		return fmt.Errorf("storage queue full (backpressure)")
	}
}

func (q *StorageQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.jobs:
			var err error
			switch job.Action {
			case "SaveBlock":
				err = q.engine.blockStore.Save(job.Data.(*Block))
			case "SaveTransaction":
				err = q.engine.txStore.Save(job.Data.(*Transaction))
			case "SaveUTXO":
				err = q.engine.utxoStore.Save(job.Data.(*UTXOEntry))
			default:
				err = fmt.Errorf("unknown storage action: %s", job.Action)
			}
			job.Done <- err
		}
	}
}
