package script

import (
	"context"
	"fmt"
	"sync"
)

type ScriptJob struct {
	ScriptSig    []byte
	ScriptPubKey []byte
	TxHash       []byte
	ResultChan   chan error
}

type ScriptQueue struct {
	jobs    chan *ScriptJob
	engine  *ScriptEngine
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewScriptQueue(size int, workers int, engine *ScriptEngine) *ScriptQueue {
	return &ScriptQueue{
		jobs:    make(chan *ScriptJob, size),
		workers: workers,
		engine:  engine,
	}
}

func (q *ScriptQueue) Start(ctx context.Context) {
	q.ctx, q.cancel = context.WithCancel(ctx)
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *ScriptQueue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
	q.wg.Wait()
}

func (q *ScriptQueue) Enqueue(job *ScriptJob) error {
	select {
	case <-q.ctx.Done():
		return fmt.Errorf("script queue stopped")
	case q.jobs <- job:
		return nil
	default:
		return fmt.Errorf("script queue full (backpressure)")
	}
}

func (q *ScriptQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.jobs:
			err := q.engine.executeInternal(job.ScriptSig, job.ScriptPubKey, job.TxHash)
			job.ResultChan <- err
		}
	}
}
