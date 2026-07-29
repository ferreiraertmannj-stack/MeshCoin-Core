package mining

import (
	"context"
	"fmt"
	"sync"
)

type TriggerEvent struct {
	Reason string
}

type TemplateQueue struct {
	jobs   chan TriggerEvent
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewTemplateQueue(size int) *TemplateQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &TemplateQueue{
		jobs:   make(chan TriggerEvent, size),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (q *TemplateQueue) Stop() {
	q.cancel()
	q.wg.Wait()
}

func (q *TemplateQueue) Trigger(reason string) error {
	select {
	case <-q.ctx.Done():
		return fmt.Errorf("queue stopped")
	case q.jobs <- TriggerEvent{Reason: reason}:
		return nil
	default:
		// Queue full - ignore trigger, one is likely already pending
		return fmt.Errorf("trigger queue full")
	}
}
