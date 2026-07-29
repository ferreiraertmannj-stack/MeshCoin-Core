package utxo

import (
	"context"
	"fmt"
	"sync"
)

type Block struct {
	Height       uint64
	Hash         string
	Transactions []Transaction
}

type queueJob struct {
	block *Block
	isAdd bool // true for AddBlock, false for RollbackBlock
	done  chan error
}

type UTXOQueue struct {
	jobs    chan queueJob
	engine  *UTXOEngine
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewUTXOQueue(size int, workers int, engine *UTXOEngine) *UTXOQueue {
	return &UTXOQueue{
		jobs:    make(chan queueJob, size),
		workers: workers,
		engine:  engine,
	}
}

func (q *UTXOQueue) Start(ctx context.Context) {
	q.ctx, q.cancel = context.WithCancel(ctx)
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *UTXOQueue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
	q.wg.Wait()
}

func (q *UTXOQueue) Enqueue(block *Block, isAdd bool) error {
	job := queueJob{
		block: block,
		isAdd: isAdd,
		done:  make(chan error, 1),
	}

	select {
	case <-q.ctx.Done():
		return fmt.Errorf("queue stopped")
	case q.jobs <- job:
		// Wait for synchronous execution if needed, but in this engine
		// the queue is just to serialize block operations safely.
		// Wait, the requirement says "fila assíncrona... sem busy waiting".
		// But AddBlock should probably know if it failed.
		// I'll return the channel error blockingly for safety, or we can just fire and forget.
		// Let's block here to ensure caller knows if validation failed.
		err := <-job.done
		return err
	default:
		return fmt.Errorf("queue full (backpressure)")
	}
}

func (q *UTXOQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.jobs:
			if job.isAdd {
				err := q.engine.processAddBlock(job.block)
				job.done <- err
			} else {
				err := q.engine.processRollbackBlock(job.block)
				job.done <- err
			}
		}
	}
}
