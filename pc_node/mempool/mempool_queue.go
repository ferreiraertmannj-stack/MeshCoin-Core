package mempool

import (
	"context"
	"fmt"
	"sync"
)

type MempoolQueue struct {
	jobs    chan Transaction
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mempool *TransactionPool

	workers int
}

func NewMempoolQueue(pool *TransactionPool, maxSize int, workers int) *MempoolQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &MempoolQueue{
		jobs:    make(chan Transaction, maxSize),
		ctx:     ctx,
		cancel:  cancel,
		mempool: pool,
		workers: workers,
	}
}

func (q *MempoolQueue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *MempoolQueue) Stop() {
	q.cancel()
	q.wg.Wait()
}

func (q *MempoolQueue) Enqueue(tx Transaction) error {
	select {
	case <-q.ctx.Done():
		return fmt.Errorf("queue is stopped")
	case q.jobs <- tx:
		return nil
	default:
		// Queue full - backpressure
		if q.mempool.events.OnPoolOverflow != nil {
			go q.mempool.events.OnPoolOverflow()
		}
		return fmt.Errorf("mempool queue overflow")
	}
}

func (q *MempoolQueue) worker() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case tx := <-q.jobs:
			// Just proxy to internal add which validates synchronously inside this worker
			_ = q.mempool.processTransaction(tx)
		}
	}
}
