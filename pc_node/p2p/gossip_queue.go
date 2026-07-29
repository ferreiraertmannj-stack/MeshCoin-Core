package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type GossipJob struct {
	Envelope GossipEnvelope
	Peer     *Peer // Peer who sent it, to avoid routing back
	Retries  int
}

type GossipQueue struct {
	jobs       chan GossipJob
	maxSize    int
	maxWorkers int
	maxRetries int
	timeout    time.Duration

	manager *GossipManager // to call forward

	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.RWMutex
}

func NewGossipQueue(maxSize int, maxWorkers int, maxRetries int, timeout time.Duration, manager *GossipManager) *GossipQueue {
	return &GossipQueue{
		jobs:       make(chan GossipJob, maxSize),
		maxSize:    maxSize,
		maxWorkers: maxWorkers,
		maxRetries: maxRetries,
		timeout:    timeout,
		manager:    manager,
	}
}

func (q *GossipQueue) Start() {
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

func (q *GossipQueue) Stop() {
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

func (q *GossipQueue) Enqueue(job GossipJob) error {
	q.mu.RLock()
	isRunning := q.isRunning
	q.mu.RUnlock()

	if !isRunning {
		return fmt.Errorf("queue is not running")
	}

	select {
	case q.jobs <- job:
		return nil
	default:
		// Queue is full -> backpressure
		if q.manager != nil {
			q.manager.stats.IncQueueOverflow()
			if q.manager.events.OnQueueOverflow != nil {
				go q.manager.events.OnQueueOverflow(job.Envelope.MessageID)
			}
		}
		return fmt.Errorf("queue is full")
	}
}

func (q *GossipQueue) worker() {
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
			q.process(job)
		}
	}
}

func (q *GossipQueue) process(job GossipJob) {
	err := q.manager.doForward(job.Envelope, job.Peer)
	if err != nil {
		if job.Retries < q.maxRetries {
			job.Retries++
			if q.manager != nil {
				q.manager.stats.IncRetries()
			}
			// Schedule retry asynchronously to avoid blocking worker
			go func() {
				time.Sleep(q.timeout)
				_ = q.Enqueue(job)
			}()
		}
	}
}
