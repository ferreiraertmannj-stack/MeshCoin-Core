package mempool

import (
	"context"
	"fmt"
	"sync"
)

type TransactionQueue struct {
	jobs    chan *MsgTransaction
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	workers int

	pipeline *TransactionValidationPipeline
}

func NewTransactionQueue(size int, workers int, pipeline *TransactionValidationPipeline) *TransactionQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &TransactionQueue{
		jobs:     make(chan *MsgTransaction, size),
		ctx:      ctx,
		cancel:   cancel,
		workers:  workers,
		pipeline: pipeline,
	}
}

func (q *TransactionQueue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *TransactionQueue) Stop() {
	q.cancel()
	q.wg.Wait()
}

// Enqueue adds a tx to be processed. Backpressure via non-blocking channel send.
func (q *TransactionQueue) Enqueue(tx *MsgTransaction) error {
	select {
	case <-q.ctx.Done():
		return fmt.Errorf("queue stopped")
	case q.jobs <- tx:
		return nil
	default:
		// Queue full
		q.pipeline.stats.IncQueueOverflows()
		return fmt.Errorf("transaction queue overflow")
	}
}

func (q *TransactionQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case tx := <-q.jobs:
			q.pipeline.Process(tx)
		}
	}
}
