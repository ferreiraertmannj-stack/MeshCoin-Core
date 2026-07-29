package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type InventoryJob struct {
	Item    InventoryItem
	Peer    *Peer
	Retries int
}

type InventoryQueue struct {
	jobs       chan InventoryJob
	maxSize    int
	maxWorkers int
	maxRetries int
	timeout    time.Duration

	manager *InventoryManager

	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.RWMutex
}

func NewInventoryQueue(maxSize int, maxWorkers int, maxRetries int, timeout time.Duration, manager *InventoryManager) *InventoryQueue {
	return &InventoryQueue{
		jobs:       make(chan InventoryJob, maxSize),
		maxSize:    maxSize,
		maxWorkers: maxWorkers,
		maxRetries: maxRetries,
		timeout:    timeout,
		manager:    manager,
	}
}

func (q *InventoryQueue) Start() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isRunning {
		return
	}

	q.ctx, q.cancel = context.WithCancel(context.Background())
	q.isRunning = true

	for i := 0; i < q.maxWorkers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *InventoryQueue) Stop() {
	q.mu.Lock()
	if !q.isRunning {
		q.mu.Unlock()
		return
	}
	q.cancel()
	q.isRunning = false
	q.mu.Unlock()

	q.wg.Wait()
}

func (q *InventoryQueue) Enqueue(job InventoryJob) error {
	q.mu.RLock()
	isRunning := q.isRunning
	q.mu.RUnlock()

	if !isRunning {
		return fmt.Errorf("queue is not running")
	}

	select {
	case q.jobs <- job:
		if q.manager != nil {
			q.manager.stats.AddPendingRequest()
		}
		return nil
	default:
		if q.manager != nil {
			q.manager.stats.IncQueueOverflows()
			if q.manager.events.OnQueueOverflow != nil {
				go q.manager.events.OnQueueOverflow()
			}
		}
		return fmt.Errorf("queue is full")
	}
}

func (q *InventoryQueue) worker() {
	defer q.wg.Done()

	if q.manager != nil {
		q.manager.stats.AddWorker()
		defer q.manager.stats.RemoveWorker()
	}

	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.jobs:
			if q.manager != nil {
				q.manager.stats.RemovePendingRequest()
			}
			q.process(job)
		}
	}
}

func (q *InventoryQueue) process(job InventoryJob) {
	// A request sends a MsgGetData to the peer for this single object
	// The response MsgData is awaited, but we shouldn't block the worker indefinitely.
	// We'll dispatch the message. If we don't get the item in time, we might retry.
	// Since MsgData comes as a separate packet, we can track pending requests in the Manager.
	// Here, we just execute the send.

	err := q.manager.doRequestObject(job)
	if err != nil {
		if job.Retries < q.maxRetries {
			job.Retries++
			if q.manager != nil {
				q.manager.stats.IncRetries()
			}
			go func() {
				time.Sleep(q.timeout) // Backoff
				_ = q.Enqueue(job)
			}()
		} else {
			if q.manager != nil {
				q.manager.stats.IncTimeouts()
				if q.manager.events.OnObjectTimeout != nil {
					go q.manager.events.OnObjectTimeout(job.Peer.Info().NodeID, job.Item.ObjectHash)
				}
			}
		}
	}
}
