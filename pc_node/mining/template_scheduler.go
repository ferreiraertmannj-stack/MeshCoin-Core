package mining

import (
	"time"
)

type TemplateScheduler struct {
	engine   *BlockTemplateEngine
	queue    *TemplateQueue
	interval time.Duration
}

func NewTemplateScheduler(engine *BlockTemplateEngine, queue *TemplateQueue, interval time.Duration) *TemplateScheduler {
	return &TemplateScheduler{
		engine:   engine,
		queue:    queue,
		interval: interval,
	}
}

func (s *TemplateScheduler) Start() {
	s.queue.wg.Add(2)
	go s.worker()
	go s.timer()
}

func (s *TemplateScheduler) Stop() {
	s.queue.Stop()
}

// TriggerEvent can be called from outside (e.g. by Mempool when tx arrives)
func (s *TemplateScheduler) TriggerEvent(reason string) {
	_ = s.queue.Trigger(reason)
}

func (s *TemplateScheduler) worker() {
	defer s.queue.wg.Done()
	for {
		select {
		case <-s.queue.ctx.Done():
			return
		case <-s.queue.jobs:
			// A trigger arrived, rebuild block immediately
			_, _ = s.engine.BuildNewTemplate()

			if s.engine.events.OnTemplateUpdated != nil {
				go s.engine.events.OnTemplateUpdated(s.engine.blockchain.GetHighestBlockHeight() + 1)
			}
		}
	}
}

func (s *TemplateScheduler) timer() {
	defer s.queue.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.queue.ctx.Done():
			return
		case <-ticker.C:
			// Time-based periodic trigger
			s.TriggerEvent("periodic")
		}
	}
}
